package service

import (
	"context"
	"errors"

	"oss-backend/internal/model/entity"

	"gorm.io/gorm"
)

// UserRequest 用户请求结构
type UserRequest struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password,omitempty" binding:"omitempty,min=6"`
	Avatar   string `json:"avatar,omitempty"`
}

// UserResponse 用户响应结构
type UserResponse struct {
	ID        uint   `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	Avatar    string `json:"avatar"`
	Role      string `json:"role"`
	CreatedAt string `json:"created_at"`
}

// ChangePasswordRequest 修改密码请求
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

// 定义错误
var (
	ErrOldPasswordIncorrect = errors.New("原密码不正确")
	ErrUserNotExist         = errors.New("用户不存在")
)

// UserService 用户服务接口
type UserService interface {
	// 获取用户信息
	GetUserInfo(ctx context.Context, id uint) (*UserResponse, error)
	// 更新用户信息
	UpdateUserInfo(ctx context.Context, id uint, req UserRequest) (*UserResponse, error)
	// 修改密码
	UpdatePassword(ctx context.Context, id uint, req ChangePasswordRequest) error
}

// userService 用户服务实现
type userService struct {
	db *gorm.DB
}

// NewUserService 创建用户服务
func NewUserService(db *gorm.DB) UserService {
	return &userService{
		db: db,
	}
}

// GetUserInfo 获取用户信息
func (s *userService) GetUserInfo(ctx context.Context, id uint) (*UserResponse, error) {
	var user entity.User
	if err := s.db.Preload("Role").First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotExist
		}
		return nil, err
	}

	return &UserResponse{
		ID:        uint(user.ID),
		Name:      user.Name,
		Email:     user.Email,
		Avatar:    user.Avatar,
		Role:      user.Role.Name,
		CreatedAt: user.CreatedAt.Format("2006-01-02 15:04:05"),
	}, nil
}

// UpdateUserInfo 更新用户信息
func (s *userService) UpdateUserInfo(ctx context.Context, id uint, req UserRequest) (*UserResponse, error) {
	var user entity.User
	if err := s.db.First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotExist
		}
		return nil, err
	}

	// 更新用户信息
	user.Name = req.Name
	if req.Avatar != "" {
		user.Avatar = req.Avatar
	}

	// 保存更新
	if err := s.db.Save(&user).Error; err != nil {
		return nil, err
	}

	return s.GetUserInfo(ctx, id)
}

// UpdatePassword 修改密码
func (s *userService) UpdatePassword(ctx context.Context, id uint, req ChangePasswordRequest) error {
	var user entity.User
	if err := s.db.First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrUserNotExist
		}
		return err
	}

	// 验证旧密码
	if !user.CheckPassword(req.OldPassword) {
		return ErrOldPasswordIncorrect
	}

	// 设置新密码
	if err := user.SetPassword(req.NewPassword); err != nil {
		return err
	}

	// 保存更新
	return s.db.Model(&user).Update("password_hash", user.PasswordHash).Error
}
