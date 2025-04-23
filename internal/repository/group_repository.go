package repository

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"oss-backend/internal/model/dto"
	"oss-backend/internal/model/entity"
)

// GroupRepository 群组仓库接口
type GroupRepository interface {
	// 群组管理
	CreateGroup(ctx context.Context, group *entity.Group) error
	GetGroupByID(ctx context.Context, id uint64) (*entity.Group, error)
	GetGroupByKey(ctx context.Context, key string) (*entity.Group, error)
	GetGroupByInviteCode(ctx context.Context, code string) (*entity.Group, error)
	UpdateGroup(ctx context.Context, group *entity.Group) error
	ListGroups(ctx context.Context, req *dto.GroupListRequest) ([]entity.Group, int64, error)

	// 成员管理
	AddMember(ctx context.Context, member *entity.GroupMember) error
	GetMember(ctx context.Context, groupID, userID uint64) (*entity.GroupMember, error)
	UpdateMember(ctx context.Context, member *entity.GroupMember) error
	RemoveMember(ctx context.Context, groupID, userID uint64) error
	ListMembers(ctx context.Context, groupID uint64, page, size int) ([]entity.GroupMember, int64, error)

	// 统计相关
	GetUserGroups(ctx context.Context, userID uint64) ([]entity.Group, error)
	GetMemberCount(ctx context.Context, groupID uint64) (int, error)
	GetProjectCount(ctx context.Context, groupID uint64) (int, error)
	GetStorageUsed(ctx context.Context, groupID uint64) (int64, error)

	// 邀请码管理
	GenerateInviteCode(ctx context.Context, groupID uint64, expireDays int) (string, time.Time, error)
	UpdateGroupInviteCode(ctx context.Context, groupID uint64, code string, expireAt *time.Time) error
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
	return r.db.WithContext(ctx).Create(group).Error
}

// GetGroupByID 根据ID获取群组
func (r *groupRepository) GetGroupByID(ctx context.Context, id uint64) (*entity.Group, error) {
	var group entity.Group
	err := r.db.WithContext(ctx).First(&group, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("群组不存在")
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
			return nil, fmt.Errorf("邀请码无效或已过期")
		}
		return nil, err
	}

	// 检查邀请码是否过期
	if group.InviteExpiresAt != nil && group.InviteExpiresAt.Before(time.Now()) {
		return nil, fmt.Errorf("邀请码已过期")
	}

	return &group, nil
}

// UpdateGroup 更新群组
func (r *groupRepository) UpdateGroup(ctx context.Context, group *entity.Group) error {
	return r.db.WithContext(ctx).Save(group).Error
}

// ListGroups 获取群组列表
func (r *groupRepository) ListGroups(ctx context.Context, req *dto.GroupListRequest) ([]entity.Group, int64, error) {
	db := r.db.WithContext(ctx).Model(&entity.Group{})

	// 条件查询
	if req.Name != "" {
		db = db.Where("name LIKE ?", "%"+req.Name+"%")
	}
	if req.Status > 0 {
		db = db.Where("status = ?", req.Status)
	}

	// 计算总数
	var total int64
	err := db.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// 分页查询
	var groups []entity.Group
	err = db.Preload("Creator").
		Offset((req.Page - 1) * req.Size).
		Limit(req.Size).
		Order("created_at DESC").
		Find(&groups).Error
	if err != nil {
		return nil, 0, err
	}

	return groups, total, nil
}

// AddMember 添加成员
func (r *groupRepository) AddMember(ctx context.Context, member *entity.GroupMember) error {
	return r.db.WithContext(ctx).Create(member).Error
}

// GetMember 获取成员
func (r *groupRepository) GetMember(ctx context.Context, groupID, userID uint64) (*entity.GroupMember, error) {
	var member entity.GroupMember
	err := r.db.WithContext(ctx).
		Where("group_id = ? AND user_id = ?", groupID, userID).
		First(&member).Error

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
func (r *groupRepository) RemoveMember(ctx context.Context, groupID, userID uint64) error {
	return r.db.WithContext(ctx).
		Where("group_id = ? AND user_id = ?", groupID, userID).
		Delete(&entity.GroupMember{}).Error
}

// ListMembers 获取成员列表
func (r *groupRepository) ListMembers(ctx context.Context, groupID uint64, page, size int) ([]entity.GroupMember, int64, error) {
	db := r.db.WithContext(ctx).
		Where("group_id = ?", groupID)

	// 计算总数
	var total int64
	err := db.Model(&entity.GroupMember{}).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// 分页查询
	var members []entity.GroupMember
	err = db.Preload("User").
		Offset((page - 1) * size).
		Limit(size).
		Order("joined_at DESC").
		Find(&members).Error
	if err != nil {
		return nil, 0, err
	}

	return members, total, nil
}

// GetUserGroups 获取用户所属的群组
func (r *groupRepository) GetUserGroups(ctx context.Context, userID uint64) ([]entity.Group, error) {
	var groups []entity.Group
	err := r.db.WithContext(ctx).
		Joins("JOIN group_members ON groups.id = group_members.group_id").
		Where("group_members.user_id = ?", userID).
		Preload("Creator").
		Find(&groups).Error

	return groups, err
}

// GetMemberCount 获取群组成员数量
func (r *groupRepository) GetMemberCount(ctx context.Context, groupID uint64) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&entity.GroupMember{}).
		Where("group_id = ?", groupID).
		Count(&count).Error

	return int(count), err
}

// GetProjectCount 获取群组项目数量
func (r *groupRepository) GetProjectCount(ctx context.Context, groupID uint64) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&entity.Project{}).
		Where("group_id = ?", groupID).
		Count(&count).Error

	return int(count), err
}

// GetStorageUsed 获取群组存储使用量
func (r *groupRepository) GetStorageUsed(ctx context.Context, groupID uint64) (int64, error) {
	// 此处需要根据文件大小统计存储量
	// 由于文件系统尚未实现，暂时返回0
	return 0, nil
}

// GenerateInviteCode 生成邀请码
func (r *groupRepository) GenerateInviteCode(ctx context.Context, groupID uint64, expireDays int) (string, time.Time, error) {
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
func (r *groupRepository) UpdateGroupInviteCode(ctx context.Context, groupID uint64, code string, expireAt *time.Time) error {
	return r.db.WithContext(ctx).
		Model(&entity.Group{}).
		Where("id = ?", groupID).
		Updates(map[string]interface{}{
			"invite_code":       code,
			"invite_expires_at": expireAt,
		}).Error
}
