package repository

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"gorm.io/gorm"

	"oss-backend/internal/model/dto"
	"oss-backend/internal/model/entity"
	"oss-backend/internal/utils"
)

// GroupRepository 群组仓库接口
type GroupRepository interface {
	// 群组管理
	CreateGroup(ctx context.Context, group *entity.Group) error
	GetGroupByID(ctx context.Context, id string) (*entity.Group, error)
	GetGroupByKey(ctx context.Context, key string) (*entity.Group, error)
	GetGroupByInviteCode(ctx context.Context, code string) (*entity.Group, error)
	UpdateGroup(ctx context.Context, group *entity.Group) error
	ListGroups(ctx context.Context, req *dto.GroupListRequest) ([]entity.Group, int64, error)

	// 成员管理
	AddMember(ctx context.Context, member *entity.GroupMember) error
	GetMember(ctx context.Context, groupID, userID string) (*entity.GroupMember, error)
	UpdateMember(ctx context.Context, member *entity.GroupMember) error
	RemoveMember(ctx context.Context, groupID, userID string) error
	ListMembers(ctx context.Context, groupID string, page, size int) ([]entity.GroupMember, int64, error)

	// 统计相关
	GetUserGroups(ctx context.Context, userID string) ([]entity.Group, error)
	GetMemberCount(ctx context.Context, groupID string) (int, error)
	GetProjectCount(ctx context.Context, groupID string) (int, error)
	GetStorageUsed(ctx context.Context, groupID string) (int64, error)

	// 邀请码管理
	GenerateInviteCode(ctx context.Context, groupID string, expireDays int) (string, time.Time, error)
	UpdateGroupInviteCode(ctx context.Context, groupID string, code string, expireAt *time.Time) error

	// 新增方法：权限检查
	CheckUserGroupRole(ctx context.Context, userID, groupID string, role string) (bool, error)
	CheckUserInGroup(ctx context.Context, userID, groupID string) (bool, error)
}

// groupRepository 群组仓库实现
type groupRepository struct {
	db *gorm.DB
}

// NewGroupRepository 创建群组仓库
func NewGroupRepository(db *gorm.DB) GroupRepository {
	return &groupRepository{
		db: db,
	}
}

// CreateGroup 创建群组
func (r *groupRepository) CreateGroup(ctx context.Context, group *entity.Group) error {
	if group.ID == "" {
		group.ID = utils.GenerateGroupID()
	}
	return r.db.WithContext(ctx).Create(group).Error
}

// GetGroupByID 根据ID获取群组
func (r *groupRepository) GetGroupByID(ctx context.Context, id string) (*entity.Group, error) {
	var group entity.Group
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&group).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &group, nil
}

// GetGroupByKey 根据Key获取群组
func (r *groupRepository) GetGroupByKey(ctx context.Context, key string) (*entity.Group, error) {
	var group entity.Group
	err := r.db.WithContext(ctx).Where("group_key = ?", key).First(&group).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &group, nil
}

// GetGroupByInviteCode 根据邀请码获取群组
func (r *groupRepository) GetGroupByInviteCode(ctx context.Context, code string) (*entity.Group, error) {
	var group entity.Group
	err := r.db.WithContext(ctx).Where("invite_code = ?", code).First(&group).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &group, nil
}

// UpdateGroup 更新群组信息
func (r *groupRepository) UpdateGroup(ctx context.Context, group *entity.Group) error {
	return r.db.WithContext(ctx).Save(group).Error
}

// ListGroups 获取群组列表
func (r *groupRepository) ListGroups(ctx context.Context, req *dto.GroupListRequest) ([]entity.Group, int64, error) {
	var groups []entity.Group
	var total int64

	query := r.db.WithContext(ctx).Model(&entity.Group{})

	// 条件筛选
	if req.Name != "" {
		query = query.Where("name LIKE ?", "%"+req.Name+"%")
	}

	if req.Status > 0 {
		query = query.Where("status = ?", req.Status)
	}

	if req.CreatorID != "" {
		query = query.Where("creator_id = ?", req.CreatorID)
	}

	// 计算总数
	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// 分页查询
	if req.Page > 0 && req.PageSize > 0 {
		offset := (req.Page - 1) * req.PageSize
		query = query.Offset(offset).Limit(req.PageSize)
	}

	// 排序
	if req.SortBy != "" {
		order := req.SortBy
		if req.SortOrder == "desc" {
			order += " DESC"
		} else {
			order += " ASC"
		}
		query = query.Order(order)
	} else {
		query = query.Order("created_at DESC")
	}

	// 执行查询
	err = query.Find(&groups).Error
	if err != nil {
		return nil, 0, err
	}

	return groups, total, nil
}

// AddMember 添加成员
func (r *groupRepository) AddMember(ctx context.Context, member *entity.GroupMember) error {
	if member.ID == "" {
		member.ID = utils.GenerateRecordID()
	}
	return r.db.WithContext(ctx).Create(member).Error
}

// GetMember 获取群组成员
func (r *groupRepository) GetMember(ctx context.Context, groupID, userID string) (*entity.GroupMember, error) {
	var member entity.GroupMember
	err := r.db.WithContext(ctx).Where("group_id = ? AND user_id = ?", groupID, userID).First(&member).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &member, nil
}

// UpdateMember 更新成员
func (r *groupRepository) UpdateMember(ctx context.Context, member *entity.GroupMember) error {
	return r.db.WithContext(ctx).Save(member).Error
}

// RemoveMember 移除成员
func (r *groupRepository) RemoveMember(ctx context.Context, groupID, userID string) error {
	return r.db.WithContext(ctx).Where("group_id = ? AND user_id = ?", groupID, userID).Delete(&entity.GroupMember{}).Error
}

// ListMembers 获取群组成员列表
func (r *groupRepository) ListMembers(ctx context.Context, groupID string, page, size int) ([]entity.GroupMember, int64, error) {
	var members []entity.GroupMember
	var total int64

	query := r.db.WithContext(ctx).Model(&entity.GroupMember{}).Where("group_id = ?", groupID)

	// 计算总数
	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// 分页查询
	if page > 0 && size > 0 {
		offset := (page - 1) * size
		query = query.Offset(offset).Limit(size)
	}

	// 执行查询，预加载用户信息
	err = query.Preload("User").Find(&members).Error
	if err != nil {
		return nil, 0, err
	}

	return members, total, nil
}

// GetUserGroups 获取用户加入的群组
func (r *groupRepository) GetUserGroups(ctx context.Context, userID string) ([]entity.Group, error) {
	var groups []entity.Group
	err := r.db.WithContext(ctx).
		Joins("JOIN group_members ON group_members.group_id = groups.id").
		Where("group_members.user_id = ?", userID).
		Find(&groups).Error
	return groups, err
}

// GetMemberCount 获取群组成员数量
func (r *groupRepository) GetMemberCount(ctx context.Context, groupID string) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entity.GroupMember{}).Where("group_id = ?", groupID).Count(&count).Error
	return int(count), err
}

// GetProjectCount 获取群组项目数量
func (r *groupRepository) GetProjectCount(ctx context.Context, groupID string) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entity.Project{}).Where("group_id = ?", groupID).Count(&count).Error
	return int(count), err
}

// GetStorageUsed 获取群组存储使用量
func (r *groupRepository) GetStorageUsed(ctx context.Context, groupID string) (int64, error) {
	type Result struct {
		TotalSize int64
	}
	var result Result
	err := r.db.WithContext(ctx).Model(&entity.File{}).
		Select("COALESCE(SUM(file_size), 0) as total_size").
		Joins("JOIN projects ON files.project_id = projects.id").
		Where("projects.group_id = ? AND files.is_deleted = ?", groupID, false).
		Where("files.is_folder = ?", false).
		Scan(&result).Error

	if err != nil {
		return 0, err
	}
	return result.TotalSize, nil
}

// GenerateInviteCode 生成邀请码
func (r *groupRepository) GenerateInviteCode(ctx context.Context, groupID string, expireDays int) (string, time.Time, error) {
	// 生成随机邀请码
	buf := make([]byte, 8)
	_, err := rand.Read(buf)
	if err != nil {
		return "", time.Time{}, err
	}

	code := hex.EncodeToString(buf)

	// 设置过期时间
	var expireAt time.Time
	if expireDays > 0 {
		expireAt = time.Now().AddDate(0, 0, expireDays)
	} else {
		// 默认30天
		expireAt = time.Now().AddDate(0, 0, 30)
	}

	// 更新到数据库
	err = r.UpdateGroupInviteCode(ctx, groupID, code, &expireAt)
	if err != nil {
		return "", time.Time{}, err
	}

	return code, expireAt, nil
}

// UpdateGroupInviteCode 更新群组邀请码
func (r *groupRepository) UpdateGroupInviteCode(ctx context.Context, groupID string, code string, expireAt *time.Time) error {
	return r.db.WithContext(ctx).
		Model(&entity.Group{}).
		Where("id = ?", groupID).
		Updates(map[string]interface{}{
			"invite_code":       code,
			"invite_expires_at": expireAt,
		}).Error
}

// CheckUserGroupRole 检查用户在群组中是否拥有指定角色
func (r *groupRepository) CheckUserGroupRole(ctx context.Context, userID, groupID string, role string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entity.GroupMember{}).
		Where("group_id = ? AND user_id = ? AND role = ?", groupID, userID, role).
		Count(&count).Error

	return count > 0, err
}

// CheckUserInGroup 检查用户是否在群组中
func (r *groupRepository) CheckUserInGroup(ctx context.Context, userID, groupID string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entity.GroupMember{}).
		Where("group_id = ? AND user_id = ?", groupID, userID).
		Count(&count).Error

	return count > 0, err
}
