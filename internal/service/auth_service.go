package service

import (
	"context"
	"errors"
	"time"

	"github.com/leoncoles/oss-backend/internal/config"
	"github.com/leoncoles/oss-backend/internal/model"
	"github.com/leoncoles/oss-backend/internal/repository"
	"github.com/leoncoles/oss-backend/internal/util"
)

// 定义错误
var (
	ErrInvalidCredentials = errors.New("邮箱或密码不正确")
	ErrUserExists         = errors.New("该邮箱已注册")
	ErrInvalidToken       = errors.New("无效的令牌")
	ErrTokenExpired       = errors.New("令牌已过期")
)

// TokenResponse 令牌响应结构
type TokenResponse struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	TokenType    string    `json:"token_type"`
}

// RegisterRequest 注册请求结构
type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Name     string `json:"name" binding:"required"`
}

// LoginRequest 登录请求结构
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// AuthService 认证服务接口
type AuthService interface {
	Register(ctx context.Context, req *RegisterRequest) (uint64, error)
	Login(ctx context.Context, req *LoginRequest, ip string) (*TokenResponse, *model.User, error)
	RefreshToken(ctx context.Context, refreshToken string) (*TokenResponse, error)
	GetUserByID(ctx context.Context, id uint64) (*model.User, error)
	ChangePassword(ctx context.Context, userID uint64, oldPassword, newPassword string) error
}

// AuthServiceImpl 认证服务实现
type AuthServiceImpl struct {
	userRepo repository.UserRepository
}

// NewAuthService 创建认证服务实例
func NewAuthService(userRepo repository.UserRepository) AuthService {
	return &AuthServiceImpl{
		userRepo: userRepo,
	}
}

// Register 用户注册
func (s *AuthServiceImpl) Register(ctx context.Context, req *RegisterRequest) (uint64, error) {
	// 创建用户对象
	user := &model.User{
		Email: req.Email,
		Name:  req.Name,
	}

	// 设置密码
	if err := user.SetPassword(req.Password); err != nil {
		return 0, err
	}

	// 保存用户
	if err := s.userRepo.Create(ctx, user); err != nil {
		if errors.Is(err, repository.ErrUserAlreadyExist) {
			return 0, ErrUserExists
		}
		return 0, err
	}

	return user.ID, nil
}

// Login 用户登录
func (s *AuthServiceImpl) Login(ctx context.Context, req *LoginRequest, ip string) (*TokenResponse, *model.User, error) {
	// 查找用户
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, nil, ErrInvalidCredentials
		}
		return nil, nil, err
	}

	// 验证密码
	if !user.CheckPassword(req.Password) {
		return nil, nil, ErrInvalidCredentials
	}

	// 更新登录信息
	if err := s.userRepo.UpdateLastLogin(ctx, user.ID, ip); err != nil {
		return nil, nil, err
	}

	// 生成访问令牌
	accessToken, err := util.GenerateToken(user.ID, user.Email, false)
	if err != nil {
		return nil, nil, err
	}

	// 生成刷新令牌
	refreshToken, err := util.GenerateToken(user.ID, user.Email, true)
	if err != nil {
		return nil, nil, err
	}

	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		return nil, nil, err
	}

	// 返回令牌信息
	return &TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Now().Add(time.Hour * time.Duration(cfg.JWT.ExpireHours)),
		TokenType:    "Bearer",
	}, user, nil
}

// RefreshToken 刷新访问令牌
func (s *AuthServiceImpl) RefreshToken(ctx context.Context, refreshToken string) (*TokenResponse, error) {
	// 解析刷新令牌
	claims, err := util.ParseToken(refreshToken)
	if err != nil {
		if errors.Is(err, util.ErrExpiredToken) {
			return nil, ErrTokenExpired
		}
		return nil, ErrInvalidToken
	}

	// 查找用户
	user, err := s.userRepo.GetByID(ctx, claims.UserID)
	if err != nil {
		return nil, err
	}

	// 生成新的访问令牌
	accessToken, err := util.GenerateToken(user.ID, user.Email, false)
	if err != nil {
		return nil, err
	}

	// 生成新的刷新令牌
	newRefreshToken, err := util.GenerateToken(user.ID, user.Email, true)
	if err != nil {
		return nil, err
	}

	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	// 返回令牌信息
	return &TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		ExpiresAt:    time.Now().Add(time.Hour * time.Duration(cfg.JWT.ExpireHours)),
		TokenType:    "Bearer",
	}, nil
}

// GetUserByID 根据ID获取用户
func (s *AuthServiceImpl) GetUserByID(ctx context.Context, id uint64) (*model.User, error) {
	return s.userRepo.GetByID(ctx, id)
}

// ChangePassword 修改密码
func (s *AuthServiceImpl) ChangePassword(ctx context.Context, userID uint64, oldPassword, newPassword string) error {
	// 获取用户
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	// 验证旧密码
	if !user.CheckPassword(oldPassword) {
		return ErrInvalidCredentials
	}

	// 更新密码
	return s.userRepo.ChangePassword(ctx, userID, newPassword)
} 