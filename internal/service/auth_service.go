package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/casbin/casbin/v2"
	"gorm.io/gorm"

	"oss-backend/internal/model/dto"
	"oss-backend/internal/model/entity"
	"oss-backend/internal/repository"
)

// 资源类型常量
const (
	ResourceProject = "projects"
	ResourceGroup   = "groups"
	ResourceFile    = "files"
	ResourceUser    = "users"
	ResourceRole    = "roles"
)

// 操作类型常量
const (
	ActionCreate = "create"
	ActionRead   = "read"
	ActionUpdate = "update"
	ActionDelete = "delete"
)

// AuthService 统一认证授权服务接口
type AuthService interface {
	// Casbin服务部分
	CheckPermission(sub, domain, obj, act string) (bool, error)
	AddRoleForUser(ctx context.Context, userID uint64, role string, domain string) error
	RemoveRoleForUser(ctx context.Context, userID uint64, role string, domain string) error
	GetRolesForUser(subject string, domain string) ([]string, error)
	InitializeRBAC() error

	// 角色服务部分
	GetAllRoles(ctx context.Context) ([]entity.Role, error)
	GetRoleByID(ctx context.Context, id uint) (*entity.Role, error)
	GetRoleByCode(ctx context.Context, code string) (*entity.Role, error)
	CreateRole(ctx context.Context, role *entity.Role) error
	UpdateRole(ctx context.Context, role *entity.Role) error
	DeleteRole(ctx context.Context, id uint) error
	ListRoles(ctx context.Context, req *dto.RoleListRequest) (*dto.RoleListResponse, error)

	// 为控制器提供DTO适配方法
	CreateRoleFromDTO(ctx context.Context, req *dto.RoleCreateRequest, createdBy uint) error
	UpdateRoleFromDTO(ctx context.Context, req *dto.RoleUpdateRequest, updatedBy uint) error

	// 用户角色关联部分
	AssignRolesToUser(ctx context.Context, userID uint64, roleIDs []uint) error
	RemoveRolesFromUser(ctx context.Context, userID uint64, roleIDs []uint) error
	GetUserRoles(ctx context.Context, userID uint64) ([]entity.Role, error)

	// 权限检查辅助方法
	CanUserAccessResource(ctx context.Context, userID uint64, resourceType, action, domain string) (bool, error)
	IsUserInRole(ctx context.Context, userID uint64, roleCode string) (bool, error)
}

// authService 认证授权服务实现
type authService struct {
	enforcer   *casbin.Enforcer
	roleRepo   repository.RoleRepository
	userRepo   repository.UserRepository
	casbinRepo repository.CasbinRepository
	db         *gorm.DB
}

// NewAuthService 创建认证授权服务
func NewAuthService(
	enforcer *casbin.Enforcer,
	roleRepo repository.RoleRepository,
	userRepo repository.UserRepository,
	casbinRepo repository.CasbinRepository,
	db *gorm.DB,
) AuthService {
	return &authService{
		enforcer:   enforcer,
		roleRepo:   roleRepo,
		userRepo:   userRepo,
		casbinRepo: casbinRepo,
		db:         db,
	}
}

// Casbin服务部分实现

// CheckPermission 检查权限
func (s *authService) CheckPermission(sub, domain, obj, act string) (bool, error) {
	return s.enforcer.Enforce(sub, domain, obj, act)
}

// AddRoleForUser 为用户添加角色
func (s *authService) AddRoleForUser(ctx context.Context, userID uint64, role string, domain string) error {
	// 构造用户标识
	sub := fmt.Sprintf("user:%d", userID)

	// 添加用户角色关联
	_, err := s.enforcer.AddRoleForUser(sub, role, domain)
	return err
}

// RemoveRoleForUser 移除用户角色
func (s *authService) RemoveRoleForUser(ctx context.Context, userID uint64, role string, domain string) error {
	// 构造用户标识
	sub := fmt.Sprintf("user:%d", userID)

	// 移除用户角色关联
	_, err := s.enforcer.DeleteRoleForUser(sub, role, domain)
	return err
}

// GetRolesForUser 获取用户角色
func (s *authService) GetRolesForUser(subject string, domain string) ([]string, error) {
	return s.enforcer.GetRolesForUser(subject, domain)
}

// InitializeRBAC 初始化RBAC策略
func (s *authService) InitializeRBAC() error {
	// 添加默认角色的权限策略
	policies := [][]string{
		// 群组管理员权限
		{"GROUP_ADMIN", "*", "projects", "create"},
		{"GROUP_ADMIN", "*", "projects", "read"},
		{"GROUP_ADMIN", "*", "projects", "update"},
		{"GROUP_ADMIN", "*", "projects", "delete"},
		{"GROUP_ADMIN", "*", "groups", "read"},
		{"GROUP_ADMIN", "*", "groups", "update"},
		{"GROUP_ADMIN", "*", "users", "read"},
		{"GROUP_ADMIN", "*", "roles", "assign"},
		{"GROUP_ADMIN", "*", "members", "add"},
		{"GROUP_ADMIN", "*", "members", "remove"},
		{"GROUP_ADMIN", "*", "files", "create"},
		{"GROUP_ADMIN", "*", "files", "read"},
		{"GROUP_ADMIN", "*", "files", "update"},
		{"GROUP_ADMIN", "*", "files", "delete"},

		// 普通成员权限
		{"MEMBER", "*", "projects", "read"},
		{"MEMBER", "*", "files", "create"},
		{"MEMBER", "*", "files", "read"},
		{"MEMBER", "*", "files", "update"},
		{"MEMBER", "*", "files", "delete"},
	}

	// 清除现有策略
	_, err := s.enforcer.RemoveFilteredPolicy(0, "GROUP_ADMIN", "MEMBER")
	if err != nil {
		return err
	}

	// 添加新策略
	_, err = s.enforcer.AddPolicies(policies)
	return err
}

// 角色服务部分实现

// GetAllRoles 获取所有角色
func (s *authService) GetAllRoles(ctx context.Context) ([]entity.Role, error) {
	return s.roleRepo.GetAll(ctx)
}

// GetRoleByID 根据ID获取角色
func (s *authService) GetRoleByID(ctx context.Context, id uint) (*entity.Role, error) {
	return s.roleRepo.GetByID(ctx, id)
}

// GetRoleByCode 根据代码获取角色
func (s *authService) GetRoleByCode(ctx context.Context, code string) (*entity.Role, error) {
	return s.roleRepo.GetByCode(ctx, code)
}

// CreateRole 创建角色
func (s *authService) CreateRole(ctx context.Context, role *entity.Role) error {
	// 检查角色代码是否已存在
	existRole, err := s.roleRepo.GetByCode(ctx, role.Code)
	if err == nil && existRole != nil {
		return errors.New("角色代码已存在")
	}

	return s.roleRepo.Create(ctx, role)
}

// UpdateRole 更新角色
func (s *authService) UpdateRole(ctx context.Context, role *entity.Role) error {
	// 检查角色是否存在
	existRole, err := s.roleRepo.GetByID(ctx, role.ID)
	if err != nil {
		return errors.New("角色不存在")
	}

	// 如果是系统角色，不允许修改代码
	if existRole.IsSystem && existRole.Code != role.Code {
		return errors.New("系统角色不允许修改代码")
	}

	return s.roleRepo.Update(ctx, role)
}

// DeleteRole 删除角色
func (s *authService) DeleteRole(ctx context.Context, id uint) error {
	// 检查角色是否存在
	role, err := s.roleRepo.GetByID(ctx, id)
	if err != nil {
		return errors.New("角色不存在")
	}

	// 系统角色不允许删除
	if role.IsSystem {
		return errors.New("系统角色不允许删除")
	}

	// 开启事务
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 删除与该角色相关的Casbin规则
		if err := s.casbinRepo.DeleteRoleRules(tx, role.Code); err != nil {
			return err
		}

		// 删除角色
		return s.roleRepo.Delete(ctx, id)
	})
}

// 用户角色关联部分实现

// AssignRolesToUser 为用户分配角色
func (s *authService) AssignRolesToUser(ctx context.Context, userID uint64, roleIDs []uint) error {
	// 获取角色信息
	var roles []entity.Role
	for _, roleID := range roleIDs {
		role, err := s.roleRepo.GetByID(ctx, roleID)
		if err != nil {
			return fmt.Errorf("角色ID %d 不存在", roleID)
		}
		roles = append(roles, *role)
	}

	// 开启事务
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 分配数据库中的用户角色关联
		err := s.userRepo.AssignRoles(ctx, userID, roleIDs)
		if err != nil {
			return err
		}

		// 同步到Casbin
		sub := fmt.Sprintf("user:%d", userID)
		for _, role := range roles {
			// 为简单起见，这里使用全局域("0")
			_, err = s.enforcer.AddRoleForUser(sub, role.Code, "0")
			if err != nil {
				return err
			}
		}

		return nil
	})
}

// RemoveRolesFromUser 移除用户角色
func (s *authService) RemoveRolesFromUser(ctx context.Context, userID uint64, roleIDs []uint) error {
	// 获取角色信息
	var roles []entity.Role
	for _, roleID := range roleIDs {
		role, err := s.roleRepo.GetByID(ctx, roleID)
		if err != nil {
			return fmt.Errorf("角色ID %d 不存在", roleID)
		}
		roles = append(roles, *role)
	}

	// 开启事务
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 移除数据库中的用户角色关联
		err := s.userRepo.RemoveRoles(ctx, userID, roleIDs)
		if err != nil {
			return err
		}

		// 同步到Casbin
		sub := fmt.Sprintf("user:%d", userID)
		for _, role := range roles {
			// 为简单起见，这里使用全局域("0")
			_, err = s.enforcer.DeleteRoleForUser(sub, role.Code, "0")
			if err != nil {
				return err
			}
		}

		return nil
	})
}

// GetUserRoles 获取用户角色
func (s *authService) GetUserRoles(ctx context.Context, userID uint64) ([]entity.Role, error) {
	return s.userRepo.GetUserRoles(ctx, userID)
}

// 权限检查辅助方法

// CanUserAccessResource 检查用户是否有权限访问资源
func (s *authService) CanUserAccessResource(ctx context.Context, userID uint64, resourceType, action, domain string) (bool, error) {
	// 构造用户标识
	sub := fmt.Sprintf("user:%d", userID)

	// 直接检查用户权限
	allowed, err := s.enforcer.Enforce(sub, domain, resourceType, action)
	if err != nil {
		return false, err
	}

	if allowed {
		return true, nil
	}

	// 检查用户角色权限
	roles, err := s.enforcer.GetRolesForUser(sub, domain)
	if err != nil {
		return false, err
	}

	for _, role := range roles {
		allowed, err := s.enforcer.Enforce(role, domain, resourceType, action)
		if err != nil {
			continue
		}
		if allowed {
			return true, nil
		}
	}

	return false, nil
}

// IsUserInRole 检查用户是否具有特定角色
func (s *authService) IsUserInRole(ctx context.Context, userID uint64, roleCode string) (bool, error) {
	// 获取用户角色
	roles, err := s.userRepo.GetUserRoles(ctx, userID)
	if err != nil {
		return false, err
	}

	// 检查用户是否拥有指定角色
	for _, role := range roles {
		if strings.EqualFold(role.Code, roleCode) {
			return true, nil
		}
	}

	return false, nil
}

// ListRoles 获取角色列表
func (s *authService) ListRoles(ctx context.Context, req *dto.RoleListRequest) (*dto.RoleListResponse, error) {
	// 默认值处理
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Size <= 0 {
		req.Size = 10
	}

	// 使用角色仓库获取角色列表
	roles, total, err := s.roleRepo.List(ctx, req.Name, req.Status, req.Page, req.Size)
	if err != nil {
		return nil, err
	}

	// 构建响应
	result := &dto.RoleListResponse{
		Total: total,
		List:  make([]dto.RoleResponse, 0, len(roles)),
	}

	for _, role := range roles {
		createdAt := role.CreatedAt.Format("2006-01-02 15:04:05")
		result.List = append(result.List, dto.RoleResponse{
			ID:          role.ID,
			Name:        role.Name,
			Description: role.Description,
			Code:        role.Code,
			Status:      role.Status,
			IsSystem:    role.IsSystem,
			CreatedAt:   createdAt,
		})
	}

	return result, nil
}

// 为控制器提供DTO适配方法

// CreateRoleFromDTO 创建角色
func (s *authService) CreateRoleFromDTO(ctx context.Context, req *dto.RoleCreateRequest, createdBy uint) error {
	// 检查角色代码是否已存在
	existRole, err := s.roleRepo.GetByCode(ctx, req.Code)
	if err == nil && existRole != nil {
		return errors.New("角色代码已存在")
	}

	role := &entity.Role{
		Name:        req.Name,
		Description: req.Description,
		Code:        req.Code,
		Status:      1,     // 默认启用
		IsSystem:    false, // 默认非系统角色
		CreatedBy:   createdBy,
	}

	return s.roleRepo.Create(ctx, role)
}

// UpdateRoleFromDTO 更新角色
func (s *authService) UpdateRoleFromDTO(ctx context.Context, req *dto.RoleUpdateRequest, updatedBy uint) error {
	// 检查角色是否存在
	existRole, err := s.roleRepo.GetByID(ctx, req.ID)
	if err != nil {
		return errors.New("角色不存在")
	}

	// 系统角色不允许修改代码（由于UpdateRequest中没有Code字段，系统角色直接保持原有代码）

	role := &entity.Role{
		ID:          req.ID,
		Name:        req.Name,
		Description: req.Description,
		Code:        existRole.Code, // 使用现有角色的Code
		Status:      req.Status,
		IsSystem:    existRole.IsSystem, // 使用现有角色的IsSystem
		UpdatedBy:   updatedBy,
	}

	return s.roleRepo.Update(ctx, role)
}
