package repository

import (
	"context"
	"errors"

	"github.com/leoncoles/oss-backend/internal/model"
	"gorm.io/gorm"
)

// 定义错误
var (
	ErrUserNotFound     = errors.New("用户不存在")
	ErrUserAlreadyExist = errors.New("用户已存在")
	ErrInvalidPassword  = errors.New("密码不正确")
)

// UserRepository 用户仓库接口
type UserRepository interface {
	Repository
	Create(ctx context.Context, user *model.User) error
	GetByID(ctx context.Context, id uint64) (*model.User, error)
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	Update(ctx context.Context, user *model.User) error
	UpdateLastLogin(ctx context.Context, userID uint64, ip string) error
	ChangePassword(ctx context.Context, userID uint64, newPassword string) error
	List(ctx context.Context, page, size int) ([]*model.User, int64, error)
}

// UserRepositoryImpl 用户仓库实现
type UserRepositoryImpl struct {
	*BaseRepository
}

// NewUserRepository 创建用户仓库
func NewUserRepository(db *gorm.DB) UserRepository {
	return &UserRepositoryImpl{
		BaseRepository: NewBaseRepository(db),
	}
}

// Create 创建用户
func (r *UserRepositoryImpl) Create(ctx context.Context, user *model.User) error {
	// 检查邮箱是否已存在
	var count int64
	if err := r.db.Model(&model.User{}).Where("email = ?", user.Email).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return ErrUserAlreadyExist
	}
	
	return r.db.Create(user).Error
}

// GetByID 根据ID获取用户
func (r *UserRepositoryImpl) GetByID(ctx context.Context, id uint64) (*model.User, error) {
	var user model.User
	err := r.db.First(&user, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

// GetByEmail 根据邮箱获取用户
func (r *UserRepositoryImpl) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	var user model.User
	err := r.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

// Update 更新用户信息
func (r *UserRepositoryImpl) Update(ctx context.Context, user *model.User) error {
	return r.db.Save(user).Error
}

// UpdateLastLogin 更新用户最后登录信息
func (r *UserRepositoryImpl) UpdateLastLogin(ctx context.Context, userID uint64, ip string) error {
	return r.db.Model(&model.User{}).Where("id = ?", userID).Updates(map[string]interface{}{
		"last_login_at": gorm.Expr("NOW()"),
		"last_login_ip": ip,
	}).Error
}

// ChangePassword 修改用户密码
func (r *UserRepositoryImpl) ChangePassword(ctx context.Context, userID uint64, newPassword string) error {
	user, err := r.GetByID(ctx, userID)
	if err != nil {
		return err
	}
	
	// 设置新密码
	if err := user.SetPassword(newPassword); err != nil {
		return err
	}
	
	return r.db.Model(&model.User{}).Where("id = ?", userID).Update("password_hash", user.PasswordHash).Error
}

// List 获取用户列表
func (r *UserRepositoryImpl) List(ctx context.Context, page, size int) ([]*model.User, int64, error) {
	var users []*model.User
	var total int64
	
	// 计算总数
	if err := r.db.Model(&model.User{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	
	// 分页查询
	offset := (page - 1) * size
	if err := r.db.Offset(offset).Limit(size).Find(&users).Error; err != nil {
		return nil, 0, err
	}
	
	return users, total, nil
} 