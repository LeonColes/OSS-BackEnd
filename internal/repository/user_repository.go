package repository

import (
	"context"
	"errors"

	"oss-backend/internal/model/entity"

	"gorm.io/gorm"
)

// 定义错误
var (
	ErrUserNotFound     = errors.New("用户不存在")
	ErrUserAlreadyExist = errors.New("用户已存在")
)

// UserRepository 用户存储库接口
type UserRepository interface {
	// 创建用户
	Create(ctx context.Context, user *entity.User) error
	// 通过ID获取用户
	GetByID(ctx context.Context, id uint64) (*entity.User, error)
	// 通过邮箱获取用户
	GetByEmail(ctx context.Context, email string) (*entity.User, error)
	// 更新用户信息
	Update(ctx context.Context, user *entity.User) error
	// 更新最后登录信息
	UpdateLastLogin(ctx context.Context, id uint64, ip string) error
	// 修改密码
	ChangePassword(ctx context.Context, id uint64, newPassword string) error
}

// userRepository 用户存储库实现
type userRepository struct {
	db *gorm.DB
}

// NewUserRepository 创建用户存储库
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

// Create 创建用户
func (r *userRepository) Create(ctx context.Context, user *entity.User) error {
	// 检查用户是否已存在
	var count int64
	if err := r.db.Model(&entity.User{}).Where("email = ?", user.Email).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return ErrUserAlreadyExist
	}

	// 创建用户
	return r.db.Create(user).Error
}

// GetByID 通过ID获取用户
func (r *userRepository) GetByID(ctx context.Context, id uint64) (*entity.User, error) {
	var user entity.User
	if err := r.db.First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

// GetByEmail 通过邮箱获取用户
func (r *userRepository) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	var user entity.User
	if err := r.db.Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

// Update 更新用户信息
func (r *userRepository) Update(ctx context.Context, user *entity.User) error {
	return r.db.Save(user).Error
}

// UpdateLastLogin 更新最后登录信息
func (r *userRepository) UpdateLastLogin(ctx context.Context, id uint64, ip string) error {
	return r.db.Model(&entity.User{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"last_login_at": gorm.Expr("NOW()"),
			"last_login_ip": ip,
		}).Error
}

// ChangePassword 修改密码
func (r *userRepository) ChangePassword(ctx context.Context, id uint64, newPassword string) error {
	user, err := r.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if err := user.SetPassword(newPassword); err != nil {
		return err
	}

	return r.db.Model(&entity.User{}).Where("id = ?", id).
		Update("password_hash", user.PasswordHash).Error
}
