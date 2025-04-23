package repository

import (
	"context"
	"time"

	"gorm.io/gorm"

	"oss-backend/internal/model/entity"
)

// UserRepository 用户仓库接口
type UserRepository interface {
	// Create 创建用户
	Create(ctx context.Context, user *entity.User) error
	// Update 更新用户
	Update(ctx context.Context, user *entity.User) error
	// GetByID 根据ID获取用户
	GetByID(ctx context.Context, id uint64) (*entity.User, error)
	// GetByEmail 根据邮箱获取用户
	GetByEmail(ctx context.Context, email string) (*entity.User, error)
	// List 获取用户列表
	List(ctx context.Context, email, name string, status, page, size int) ([]*entity.User, int64, error)
	// UpdatePassword 更新密码
	UpdatePassword(ctx context.Context, id uint64, passwordHash string) error
	// UpdateStatus 更新状态
	UpdateStatus(ctx context.Context, id uint64, status int) error
	// UpdateLastLogin 更新最后登录信息
	UpdateLastLogin(ctx context.Context, id uint64, ip string) error
	// GetUserRoles 获取用户角色
	GetUserRoles(ctx context.Context, userID uint64) ([]entity.Role, error)
	// AssignRoles 为用户分配角色
	AssignRoles(ctx context.Context, userID uint64, roleIDs []uint) error
	// RemoveRoles 移除用户角色
	RemoveRoles(ctx context.Context, userID uint64, roleIDs []uint) error
}

// userRepository 用户仓库实现
type userRepository struct {
	db *gorm.DB
}

// NewUserRepository 创建用户仓库
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{
		db: db,
	}
}

// Create 创建用户
func (r *userRepository) Create(ctx context.Context, user *entity.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

// Update 更新用户
func (r *userRepository) Update(ctx context.Context, user *entity.User) error {
	return r.db.WithContext(ctx).Model(&entity.User{}).Where("id = ?", user.ID).
		Updates(map[string]interface{}{
			"name":       user.Name,
			"avatar":     user.Avatar,
			"updated_at": user.UpdatedAt,
		}).Error
}

// GetByID 根据ID获取用户
func (r *userRepository) GetByID(ctx context.Context, id uint64) (*entity.User, error) {
	var user entity.User
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByEmail 根据邮箱获取用户
func (r *userRepository) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	var user entity.User
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// List 获取用户列表
func (r *userRepository) List(ctx context.Context, email, name string, status, page, size int) ([]*entity.User, int64, error) {
	var users []*entity.User
	var total int64

	db := r.db.WithContext(ctx).Model(&entity.User{})

	if email != "" {
		db = db.Where("email LIKE ?", "%"+email+"%")
	}

	if name != "" {
		db = db.Where("name LIKE ?", "%"+name+"%")
	}

	if status > 0 {
		db = db.Where("status = ?", status)
	}

	err := db.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	if page > 0 && size > 0 {
		offset := (page - 1) * size
		db = db.Offset(offset).Limit(size)
	}

	err = db.Order("id DESC").Find(&users).Error
	if err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// UpdatePassword 更新密码
func (r *userRepository) UpdatePassword(ctx context.Context, id uint64, passwordHash string) error {
	return r.db.WithContext(ctx).Model(&entity.User{}).Where("id = ?", id).
		Update("password_hash", passwordHash).Error
}

// UpdateStatus 更新状态
func (r *userRepository) UpdateStatus(ctx context.Context, id uint64, status int) error {
	return r.db.WithContext(ctx).Model(&entity.User{}).Where("id = ?", id).
		Update("status", status).Error
}

// UpdateLastLogin 更新最后登录信息
func (r *userRepository) UpdateLastLogin(ctx context.Context, id uint64, ip string) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&entity.User{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"last_login_at": now,
			"last_login_ip": ip,
		}).Error
}

// GetUserRoles 获取用户角色
func (r *userRepository) GetUserRoles(ctx context.Context, userID uint64) ([]entity.Role, error) {
	var roles []entity.Role
	err := r.db.WithContext(ctx).
		Joins("JOIN user_roles ON user_roles.role_id = roles.id").
		Where("user_roles.user_id = ?", userID).
		Find(&roles).Error
	return roles, err
}

// AssignRoles 为用户分配角色
func (r *userRepository) AssignRoles(ctx context.Context, userID uint64, roleIDs []uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 创建用户角色关联
		for _, roleID := range roleIDs {
			userRole := entity.UserRole{
				UserID: uint(userID),
				RoleID: roleID,
			}
			// 检查是否已存在
			var count int64
			err := tx.Model(&entity.UserRole{}).
				Where("user_id = ? AND role_id = ?", userID, roleID).
				Count(&count).Error
			if err != nil {
				return err
			}
			// 不存在则创建
			if count == 0 {
				if err := tx.Create(&userRole).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})
}

// RemoveRoles 移除用户角色
func (r *userRepository) RemoveRoles(ctx context.Context, userID uint64, roleIDs []uint) error {
	if len(roleIDs) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Where("user_id = ? AND role_id IN ?", userID, roleIDs).
		Delete(&entity.UserRole{}).Error
}
