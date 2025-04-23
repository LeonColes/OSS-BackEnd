package service

import (
	"context"
	"errors"
	"time"

	"oss-backend/internal/config"
	"oss-backend/internal/repository"
)

// UserClaims 用户JWT声明
type UserClaims struct {
	UserID uint   `json:"user_id"`
	Email  string `json:"email"`
}

// 定义错误
var (
	ErrInvalidCredentials = errors.New("邮箱或密码不正确")
	ErrUserExists         = errors.New("该邮箱已注册")
	ErrInvalidToken       = errors.New("无效的令牌")
	ErrTokenExpired       = errors.New("令牌已过期")
)

// TokenResponse 令牌响应结构
type TokenResponse struct {
	AccessToken  string    `json:"access_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	RefreshToken string    `json:"refresh_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	ExpiresAt    time.Time `json:"expires_at" example:"2023-01-01T00:00:00Z"`
	TokenType    string    `json:"token_type" example:"Bearer"`
}

// RegisterRequest 注册请求结构
type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email" example:"user@example.com"`
	Password string `json:"password" binding:"required,min=6" example:"password123"`
	Name     string `json:"name" binding:"required" example:"张三"`
}

// LoginRequest 登录请求结构
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email" example:"admin@example.com"`
	Password string `json:"password" binding:"required" example:"admin123"`
}

// AuthService 认证服务接口
type AuthService interface {
	// 登录
	Login(ctx context.Context, req LoginRequest) (*TokenResponse, error)
	// 注册
	Register(ctx context.Context, req RegisterRequest) error
	// 刷新令牌
	RefreshToken(ctx context.Context, refreshToken string) (*TokenResponse, error)
	// 验证令牌
	ValidateToken(tokenString string) (*UserClaims, error)
}

// authService 认证服务实现
type authService struct {
	userRepo repository.UserRepository
	config   *config.Config
}

// NewAuthService 创建认证服务
func NewAuthService(userRepo repository.UserRepository) AuthService {
	return &authService{
		userRepo: userRepo,
	}
}

// Login 登录
func (s *authService) Login(ctx context.Context, req LoginRequest) (*TokenResponse, error) {
	// 功能尚未实现
	return nil, errors.New("功能尚未实现")
}

// Register 注册
func (s *authService) Register(ctx context.Context, req RegisterRequest) error {
	// 功能尚未实现
	return errors.New("功能尚未实现")
}

// RefreshToken 刷新令牌
func (s *authService) RefreshToken(ctx context.Context, refreshToken string) (*TokenResponse, error) {
	// 功能尚未实现
	return nil, errors.New("功能尚未实现")
}

// ValidateToken 验证令牌
func (s *authService) ValidateToken(tokenString string) (*UserClaims, error) {
	// 功能尚未实现
	return nil, errors.New("功能尚未实现")
}
