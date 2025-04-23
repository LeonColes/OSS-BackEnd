package service

import (
	"errors"

	"oss-backend/internal/model/entity"

	"gorm.io/gorm"
)

// 定义错误
var (
	ErrGroupNotFound          = errors.New("群组不存在")
	ErrGroupNameExists        = errors.New("群组名称已存在")
	ErrNotGroupMember         = errors.New("不是群组成员")
	ErrInsufficientPermission = errors.New("权限不足")
	ErrCannotRemoveOwner      = errors.New("无法移除群组拥有者")
	ErrGroupNotEmpty          = errors.New("群组不为空，请先删除群组中的项目")
	// ErrNoPermission 已在file_service.go中定义
)

// CreateGroupRequest 创建群组请求结构
type CreateGroupRequest struct {
	Name         string `json:"name" binding:"required" example:"研发部"`
	Description  string `json:"description" example:"技术研发部门"`
	GroupKey     string `json:"group_key" binding:"required" example:"dev-group"`
	InviteCode   string `json:"invite_code" binding:"required" example:"dev123"`
	StorageQuota int64  `json:"storage_quota" example:"10737418240"` // 默认10GB
}

// UpdateGroupRequest 更新群组请求结构
type UpdateGroupRequest struct {
	Name         string `json:"name" example:"研发部"`
	Description  string `json:"description" example:"技术研发部门"`
	InviteCode   string `json:"invite_code" example:"dev123"`
	StorageQuota int64  `json:"storage_quota" example:"10737418240"` // 默认10GB
}

// AddMemberRequest 添加成员请求结构
type AddMemberRequest struct {
	UserID uint64 `json:"user_id" binding:"required" example:"2"`
	Role   string `json:"role" binding:"required" example:"member"`
}

// GroupService 群组服务接口
type GroupService interface {
	// 创建群组
	CreateGroup(name string, description string, creatorID uint) (*entity.Group, error)

	// 获取群组信息
	GetGroupByID(groupID uint) (*entity.Group, error)

	// 更新群组信息
	UpdateGroup(groupID uint, name string, description string) (*entity.Group, error)

	// 删除群组
	DeleteGroup(groupID uint) error

	// 获取用户的群组列表
	GetUserGroups(userID uint) ([]*entity.Group, error)

	// 获取群组成员列表
	GetGroupMembers(groupID uint) ([]*entity.GroupMember, error)

	// 添加群组成员
	AddGroupMember(groupID uint, userID uint, role string) error

	// 移除群组成员
	RemoveGroupMember(groupID uint, userID uint) error

	// 更新群组成员角色
	UpdateMemberRole(groupID uint, userID uint, role string) error

	// 检查用户是否具有指定群组的指定权限
	CheckPermission(groupID uint, userID uint, requiredRoles []string) (bool, error)
}

type groupService struct {
	db *gorm.DB
}

// NewGroupService 创建群组服务实例
func NewGroupService(db *gorm.DB) GroupService {
	return &groupService{db: db}
}

// CreateGroup 创建群组
func (s *groupService) CreateGroup(name string, description string, creatorID uint) (*entity.Group, error) {
	// 检查群组名是否已存在
	var count int64
	if err := s.db.Model(&entity.Group{}).Where("name = ?", name).Count(&count).Error; err != nil {
		return nil, err
	}

	if count > 0 {
		return nil, ErrGroupNameExists
	}

	// 创建群组
	group := &entity.Group{
		Name:        name,
		Description: description,
		CreatorID:   uint64(creatorID),
		Status:      1, // 正常状态
	}

	if err := s.db.Create(group).Error; err != nil {
		return nil, err
	}

	// 将创建者添加为群组成员并设置为拥有者
	member := &entity.GroupMember{
		GroupID: group.ID,
		UserID:  uint64(creatorID),
		Role:    entity.GroupRoleOwner,
	}

	if err := s.db.Create(member).Error; err != nil {
		return nil, err
	}

	return group, nil
}

// GetGroupByID 获取群组信息
func (s *groupService) GetGroupByID(groupID uint) (*entity.Group, error) {
	group := &entity.Group{}
	if err := s.db.Preload("Creator").First(group, groupID).Error; err != nil {
		return nil, ErrGroupNotFound
	}

	return group, nil
}

// UpdateGroup 更新群组信息
func (s *groupService) UpdateGroup(groupID uint, name string, description string) (*entity.Group, error) {
	group, err := s.GetGroupByID(groupID)
	if err != nil {
		return nil, err
	}

	// 如果名称变更了，需要检查是否与其他群组重名
	if name != group.Name {
		var count int64
		if err := s.db.Model(&entity.Group{}).Where("name = ? AND id != ?", name, groupID).Count(&count).Error; err != nil {
			return nil, err
		}

		if count > 0 {
			return nil, ErrGroupNameExists
		}
	}

	// 更新群组信息
	group.Name = name
	group.Description = description

	if err := s.db.Save(group).Error; err != nil {
		return nil, err
	}

	return group, nil
}

// DeleteGroup 删除群组
func (s *groupService) DeleteGroup(groupID uint) error {
	// 检查群组是否存在
	group, err := s.GetGroupByID(groupID)
	if err != nil {
		return err
	}

	// 检查群组是否有关联的项目
	var projectCount int64
	if err := s.db.Model(&entity.Project{}).Where("group_id = ?", groupID).Count(&projectCount).Error; err != nil {
		return err
	}

	if projectCount > 0 {
		return ErrGroupNotEmpty
	}

	// 使用事务删除群组和相关成员
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 删除群组成员
		if err := tx.Where("group_id = ?", groupID).Delete(&entity.GroupMember{}).Error; err != nil {
			return err
		}

		// 删除群组
		if err := tx.Delete(group).Error; err != nil {
			return err
		}

		return nil
	})
}

// GetUserGroups 获取用户的群组列表
func (s *groupService) GetUserGroups(userID uint) ([]*entity.Group, error) {
	var groupIDs []uint

	// 获取用户所在的群组ID列表
	if err := s.db.Model(&entity.GroupMember{}).
		Where("user_id = ?", userID).
		Pluck("group_id", &groupIDs).
		Error; err != nil {
		return nil, err
	}

	var groups []*entity.Group
	if len(groupIDs) == 0 {
		return groups, nil
	}

	// 获取群组详情
	if err := s.db.
		Preload("Creator").
		Where("id IN ?", groupIDs).
		Find(&groups).
		Error; err != nil {
		return nil, err
	}

	return groups, nil
}

// GetGroupMembers 获取群组成员列表
func (s *groupService) GetGroupMembers(groupID uint) ([]*entity.GroupMember, error) {
	// 检查群组是否存在
	if _, err := s.GetGroupByID(groupID); err != nil {
		return nil, err
	}

	var members []*entity.GroupMember

	if err := s.db.
		Preload("User").
		Where("group_id = ?", groupID).
		Find(&members).
		Error; err != nil {
		return nil, err
	}

	return members, nil
}

// AddGroupMember 添加群组成员
func (s *groupService) AddGroupMember(groupID uint, userID uint, role string) error {
	// 检查群组是否存在
	if _, err := s.GetGroupByID(groupID); err != nil {
		return err
	}

	// 检查用户是否已经是成员
	var count int64
	if err := s.db.Model(&entity.GroupMember{}).
		Where("group_id = ? AND user_id = ?", groupID, userID).
		Count(&count).
		Error; err != nil {
		return err
	}

	if count > 0 {
		// 已经是成员，更新角色
		return s.UpdateMemberRole(groupID, userID, role)
	}

	// 添加新成员
	member := &entity.GroupMember{
		GroupID: uint64(groupID),
		UserID:  uint64(userID),
		Role:    role,
	}

	return s.db.Create(member).Error
}

// RemoveGroupMember 移除群组成员
func (s *groupService) RemoveGroupMember(groupID uint, userID uint) error {
	// 检查要移除的用户是否是群组拥有者
	var member entity.GroupMember
	if err := s.db.
		Where("group_id = ? AND user_id = ?", groupID, userID).
		First(&member).
		Error; err != nil {
		return ErrNotGroupMember
	}

	// 不能移除群组拥有者
	if member.Role == entity.GroupRoleOwner {
		return ErrCannotRemoveOwner
	}

	// 删除成员
	return s.db.
		Where("group_id = ? AND user_id = ?", groupID, userID).
		Delete(&entity.GroupMember{}).
		Error
}

// UpdateMemberRole 更新群组成员角色
func (s *groupService) UpdateMemberRole(groupID uint, userID uint, role string) error {
	// 检查成员是否存在
	var member entity.GroupMember
	if err := s.db.
		Where("group_id = ? AND user_id = ?", groupID, userID).
		First(&member).
		Error; err != nil {
		return ErrNotGroupMember
	}

	// 不能修改拥有者的角色
	if member.Role == entity.GroupRoleOwner && role != entity.GroupRoleOwner {
		return ErrCannotRemoveOwner
	}

	// 更新角色
	member.Role = role
	return s.db.Save(&member).Error
}

// CheckPermission 检查用户是否具有指定群组的指定权限
func (s *groupService) CheckPermission(groupID uint, userID uint, requiredRoles []string) (bool, error) {
	var member entity.GroupMember
	if err := s.db.
		Where("group_id = ? AND user_id = ?", groupID, userID).
		First(&member).
		Error; err != nil {
		return false, ErrNotGroupMember
	}

	// 检查用户角色是否在所需角色列表中
	for _, role := range requiredRoles {
		if member.Role == role {
			return true, nil
		}
	}

	return false, nil
}
