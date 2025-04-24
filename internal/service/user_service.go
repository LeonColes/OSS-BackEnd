package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"

	"oss-backend/internal/model/dto"
	"oss-backend/internal/model/entity"
	"oss-backend/internal/repository"
)

// 定义JWT密钥
var jwtSecret = []byte("oss-backend-secret-key")

// JWTClaims 自定义JWT声明结构
type JWTClaims struct {
	UserID uint64 `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

// UserService 用户服务接口
type UserService interface {
	// Register 用户注册
	Register(ctx context.Context, req *dto.UserRegisterRequest) (*dto.UserResponse, error)
	// Login 用户登录
	Login(ctx context.Context, req *dto.UserLoginRequest, ip string) (*dto.LoginResponse, error)
	// GetUserInfo 获取用户信息
	GetUserInfo(ctx context.Context, id uint64) (*dto.UserResponse, error)
	// UpdateUserInfo 更新用户信息
	UpdateUserInfo(ctx context.Context, id uint64, req *dto.UserUpdateRequest) error
	// UpdatePassword 更新密码
	UpdatePassword(ctx context.Context, id uint64, req *dto.UserPasswordUpdateRequest) error
	// ListUsers 获取用户列表
	ListUsers(ctx context.Context, req *dto.UserListRequest) (*dto.UserListResponse, error)
	// UpdateUserStatus 更新用户状态
	UpdateUserStatus(ctx context.Context, id uint64, status int) error
	// GetUserRoles 获取用户角色
	GetUserRoles(ctx context.Context, userID uint64) ([]entity.Role, error)
	// AssignRoles 为用户分配角色
	AssignRoles(ctx context.Context, userID uint64, roleIDs []uint) error
	// RemoveRoles 移除用户角色
	RemoveRoles(ctx context.Context, userID uint64, roleIDs []uint) error
	// InitAdminUser 初始化系统管理员用户
	InitAdminUser(ctx context.Context) error
}

// userService 用户服务实现
type userService struct {
	userRepo    repository.UserRepository
	roleRepo    repository.RoleRepository
	authService AuthService
}

// NewUserService 创建用户服务
func NewUserService(userRepo repository.UserRepository, roleRepo repository.RoleRepository, authService AuthService) UserService {
	return &userService{
		userRepo:    userRepo,
		roleRepo:    roleRepo,
		authService: authService,
	}
}

// Register 用户注册
func (s *userService) Register(ctx context.Context, req *dto.UserRegisterRequest) (*dto.UserResponse, error) {
	// 检查邮箱是否已存在
	existUser, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err == nil && existUser != nil {
		return nil, errors.New("邮箱已被注册")
	}

	// 生成密码哈希
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// 创建用户
	user := &entity.User{
		Email:        req.Email,
		Name:         req.Name,
		PasswordHash: string(passwordHash),
		Status:       entity.UserStatusNormal,
	}

	// 保存用户
	err = s.userRepo.Create(ctx, user)
	if err != nil {
		return nil, err
	}

	// 为用户分配默认角色（普通成员）
	defaultRole, err := s.roleRepo.GetByCode(ctx, entity.RoleMember)
	if err == nil && defaultRole != nil {
		// 分配默认角色
		err = s.userRepo.AssignRoles(ctx, uint64(user.ID), []uint{defaultRole.ID})
		if err != nil {
			// 记录错误但不阻止注册完成
			// TODO: 记录日志
		}

		// 同步到Casbin
		if s.authService != nil {
			_ = s.authService.AddRoleForUser(ctx, uint64(user.ID), entity.RoleMember, "0")
		}
	}

	// 获取刚创建的用户完整信息（包括角色）
	createdUser, err := s.userRepo.GetByID(ctx, uint64(user.ID))
	if err != nil {
		return nil, err
	}

	// 获取用户角色
	roles, _ := s.userRepo.GetUserRoles(ctx, uint64(user.ID))
	userResponse := s.convertToUserResponse(createdUser)
	userResponse.Roles = s.convertToRoleResponses(roles)

	return userResponse, nil
}

// Login 用户登录
func (s *userService) Login(ctx context.Context, req *dto.UserLoginRequest, ip string) (*dto.LoginResponse, error) {
	// 根据邮箱获取用户
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, errors.New("用户不存在或密码错误")
	}

	// 检查用户状态
	if user.Status != entity.UserStatusNormal {
		return nil, errors.New("账号已被禁用或锁定")
	}

	// 验证密码
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if err != nil {
		return nil, errors.New("用户不存在或密码错误")
	}

	// 更新最后登录信息
	err = s.userRepo.UpdateLastLogin(ctx, uint64(user.ID), ip)
	if err != nil {
		// 非致命错误，可以继续
		// TODO: 记录日志
	}

	// 生成JWT Token
	token, refreshToken, expiresAt, err := s.generateToken(uint64(user.ID), user.Email)
	if err != nil {
		return nil, errors.New("生成令牌失败")
	}

	// 获取用户角色
	roles, _ := s.userRepo.GetUserRoles(ctx, uint64(user.ID))
	userResponse := s.convertToUserResponse(user)
	userResponse.Roles = s.convertToRoleResponses(roles)

	// 转换为响应
	return &dto.LoginResponse{
		Token:        token,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt,
		UserInfo:     *userResponse,
	}, nil
}

// generateToken 生成JWT令牌
func (s *userService) generateToken(userID uint64, email string) (string, string, int64, error) {
	// Token过期时间：24小时
	expiresAt := time.Now().Add(24 * time.Hour)

	// 创建JWT声明
	claims := JWTClaims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Subject:   email,
		},
	}

	// 生成JWT令牌
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", "", 0, err
	}

	// 生成刷新令牌，过期时间更长：7天
	refreshExpiresAt := time.Now().Add(7 * 24 * time.Hour)
	refreshClaims := JWTClaims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(refreshExpiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Subject:   email + ":refresh",
		},
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString(jwtSecret)
	if err != nil {
		return "", "", 0, err
	}

	return tokenString, refreshTokenString, expiresAt.Unix(), nil
}

// GetUserInfo 获取用户信息
func (s *userService) GetUserInfo(ctx context.Context, id uint64) (*dto.UserResponse, error) {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 获取用户角色
	roles, _ := s.userRepo.GetUserRoles(ctx, id)
	userResponse := s.convertToUserResponse(user)
	userResponse.Roles = s.convertToRoleResponses(roles)

	return userResponse, nil
}

// UpdateUserInfo 更新用户信息
func (s *userService) UpdateUserInfo(ctx context.Context, id uint64, req *dto.UserUpdateRequest) error {
	// 获取用户
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// 更新用户信息
	user.Name = req.Name
	user.Avatar = req.Avatar
	user.UpdatedAt = time.Now()

	return s.userRepo.Update(ctx, user)
}

// UpdatePassword 更新密码
func (s *userService) UpdatePassword(ctx context.Context, id uint64, req *dto.UserPasswordUpdateRequest) error {
	// 获取用户
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// 验证旧密码
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.OldPassword))
	if err != nil {
		return errors.New("原密码错误")
	}

	// 生成新密码哈希
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	return s.userRepo.UpdatePassword(ctx, id, string(passwordHash))
}

// ListUsers 获取用户列表
func (s *userService) ListUsers(ctx context.Context, req *dto.UserListRequest) (*dto.UserListResponse, error) {
	// 默认值处理
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Size <= 0 {
		req.Size = 10
	}

	// 获取用户列表
	users, total, err := s.userRepo.List(ctx, req.Email, req.Name, req.Status, req.Page, req.Size)
	if err != nil {
		return nil, err
	}

	// 转换为响应
	result := &dto.UserListResponse{
		Total: total,
		List:  make([]dto.UserResponse, 0, len(users)),
	}

	for _, user := range users {
		// 获取用户角色
		roles, _ := s.userRepo.GetUserRoles(ctx, uint64(user.ID))
		userResponse := s.convertToUserResponse(user)
		userResponse.Roles = s.convertToRoleResponses(roles)

		result.List = append(result.List, *userResponse)
	}

	return result, nil
}

// UpdateUserStatus 更新用户状态
func (s *userService) UpdateUserStatus(ctx context.Context, id uint64, status int) error {
	// 检查状态值是否有效
	if status != entity.UserStatusNormal && status != entity.UserStatusDisabled && status != entity.UserStatusLocked {
		return errors.New("无效的状态值")
	}

	return s.userRepo.UpdateStatus(ctx, id, status)
}

// GetUserRoles 获取用户角色
func (s *userService) GetUserRoles(ctx context.Context, userID uint64) ([]entity.Role, error) {
	return s.userRepo.GetUserRoles(ctx, userID)
}

// AssignRoles 为用户分配角色
func (s *userService) AssignRoles(ctx context.Context, userID uint64, roleIDs []uint) error {
	// 检查用户是否存在
	_, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return errors.New("用户不存在")
	}

	// 检查角色是否存在
	var roles []entity.Role
	for _, roleID := range roleIDs {
		role, err := s.roleRepo.GetByID(ctx, roleID)
		if err != nil {
			return errors.New("角色不存在")
		}
		roles = append(roles, *role)
	}

	// 数据库事务
	err = s.userRepo.AssignRoles(ctx, userID, roleIDs)
	if err != nil {
		return err
	}

	// 同步到Casbin
	if s.authService != nil {
		// 为用户分配角色到Casbin
		for _, role := range roles {
			_ = s.authService.AddRoleForUser(ctx, userID, role.Code, "0")
		}
	}

	return nil
}

// RemoveRoles 移除用户角色
func (s *userService) RemoveRoles(ctx context.Context, userID uint64, roleIDs []uint) error {
	// 检查用户是否存在
	_, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return errors.New("用户不存在")
	}

	// 获取要删除的角色以同步Casbin
	var rolesToRemove []entity.Role
	for _, roleID := range roleIDs {
		role, err := s.roleRepo.GetByID(ctx, roleID)
		if err == nil && role != nil {
			rolesToRemove = append(rolesToRemove, *role)
		}
	}

	// 移除数据库中的角色
	err = s.userRepo.RemoveRoles(ctx, userID, roleIDs)
	if err != nil {
		return err
	}

	// 同步到Casbin
	if s.authService != nil && len(rolesToRemove) > 0 {
		for _, role := range rolesToRemove {
			_ = s.authService.RemoveRoleForUser(ctx, userID, role.Code, "0")
		}
	}

	return nil
}

// convertToUserResponse 将用户实体转换为用户响应
func (s *userService) convertToUserResponse(user *entity.User) *dto.UserResponse {
	return &dto.UserResponse{
		ID:          uint64(user.ID),
		Email:       user.Email,
		Name:        user.Name,
		Avatar:      user.Avatar,
		Status:      user.Status,
		LastLoginAt: user.LastLoginAt,
		LastLoginIP: user.LastLoginIP,
		CreatedAt:   user.CreatedAt,
	}
}

// convertToRoleResponses 将角色实体转换为角色响应
func (s *userService) convertToRoleResponses(roles []entity.Role) []dto.RoleResponse {
	if len(roles) == 0 {
		return nil
	}

	result := make([]dto.RoleResponse, 0, len(roles))
	for _, role := range roles {
		createdAt := role.CreatedAt.Format("2006-01-02 15:04:05")
		result = append(result, dto.RoleResponse{
			ID:          uint(role.ID),
			Name:        role.Name,
			Description: role.Description,
			Code:        role.Code,
			Status:      role.Status,
			IsSystem:    role.IsSystem,
			CreatedAt:   createdAt,
		})
	}
	return result
}

// InitAdminUser 初始化系统管理员用户
func (s *userService) InitAdminUser(ctx context.Context) error {
	// 检查是否已存在管理员用户
	adminRole, err := s.roleRepo.GetByCode(ctx, entity.RoleAdmin)
	if err != nil {
		return fmt.Errorf("获取管理员角色失败: %w", err)
	}

	// 查询是否有用户拥有管理员角色
	// 这里查询所有正常状态的用户，不进行分页限制
	users, _, err := s.userRepo.List(ctx, "", "", entity.UserStatusNormal, 0, 0)
	if err != nil {
		return fmt.Errorf("查询用户失败: %w", err)
	}

	// 检查用户中是否有管理员
	hasAdmin := false
	for _, user := range users {
		roles, err := s.userRepo.GetUserRoles(ctx, uint64(user.ID))
		if err != nil {
			continue
		}

		for _, role := range roles {
			if role.Code == entity.RoleAdmin {
				hasAdmin = true
				break
			}
		}
		if hasAdmin {
			break
		}
	}

	// 如果已存在管理员用户，则不需要创建
	if hasAdmin {
		return nil
	}

	// 创建管理员用户
	adminReq := &dto.UserRegisterRequest{
		Email:    "admin@x.com",
		Password: "admin123",
		Name:     "Admin",
	}

	admin, err := s.Register(ctx, adminReq)
	if err != nil {
		return fmt.Errorf("创建管理员用户失败: %w", err)
	}

	// 分配管理员角色
	err = s.AssignRoles(ctx, admin.ID, []uint{adminRole.ID})
	if err != nil {
		return fmt.Errorf("分配管理员角色失败: %w", err)
	}

	return nil
}
