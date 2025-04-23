package service

import (
	"context"
	"time"

	"oss-backend/internal/model/dto"
	"oss-backend/internal/model/entity"
	"oss-backend/internal/repository"
)

// RoleService 角色服务接口
type RoleService interface {
	// CreateRole 创建角色
	CreateRole(ctx context.Context, req *dto.RoleCreateRequest, createdBy uint) error
	// UpdateRole 更新角色
	UpdateRole(ctx context.Context, req *dto.RoleUpdateRequest, updatedBy uint) error
	// DeleteRole 删除角色
	DeleteRole(ctx context.Context, id uint) error
	// GetRoleByID 根据ID获取角色
	GetRoleByID(ctx context.Context, id uint) (*dto.RoleResponse, error)
	// GetRoleByCode 根据Code获取角色
	GetRoleByCode(ctx context.Context, code string) (*dto.RoleResponse, error)
	// ListRoles 获取角色列表
	ListRoles(ctx context.Context, req *dto.RoleListRequest) (*dto.RoleListResponse, error)
	// InitSystemRoles 初始化系统角色
	InitSystemRoles(ctx context.Context) error
}

// roleService 角色服务实现
type roleService struct {
	roleRepo repository.RoleRepository
}

// NewRoleService 创建角色服务
func NewRoleService(roleRepo repository.RoleRepository) RoleService {
	return &roleService{
		roleRepo: roleRepo,
	}
}

// CreateRole 创建角色
func (s *roleService) CreateRole(ctx context.Context, req *dto.RoleCreateRequest, createdBy uint) error {
	// 检查角色编码是否已存在
	role, err := s.roleRepo.GetByCode(ctx, req.Code)
	if err == nil && role != nil {
		return ErrRoleCodeExists
	}

	// 创建角色
	role = &entity.Role{
		Name:        req.Name,
		Description: req.Description,
		Code:        req.Code,
		Status:      1, // 默认启用
		CreatedBy:   createdBy,
		UpdatedBy:   createdBy,
	}

	return s.roleRepo.Create(ctx, role)
}

// UpdateRole 更新角色
func (s *roleService) UpdateRole(ctx context.Context, req *dto.RoleUpdateRequest, updatedBy uint) error {
	// 检查角色是否存在
	role, err := s.roleRepo.GetByID(ctx, req.ID)
	if err != nil {
		return err
	}

	// 系统角色只允许修改名称和描述
	if role.IsSystem {
		role.Name = req.Name
		role.Description = req.Description
	} else {
		role.Name = req.Name
		role.Description = req.Description
		role.Status = req.Status
	}

	role.UpdatedBy = updatedBy
	role.UpdatedAt = time.Now()

	return s.roleRepo.Update(ctx, role)
}

// DeleteRole 删除角色
func (s *roleService) DeleteRole(ctx context.Context, id uint) error {
	// 检查角色是否存在
	role, err := s.roleRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// 系统角色不允许删除
	if role.IsSystem {
		return ErrSystemRoleCannotDelete
	}

	return s.roleRepo.Delete(ctx, id)
}

// GetRoleByID 根据ID获取角色
func (s *roleService) GetRoleByID(ctx context.Context, id uint) (*dto.RoleResponse, error) {
	role, err := s.roleRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return s.convertToRoleResponse(role), nil
}

// GetRoleByCode 根据Code获取角色
func (s *roleService) GetRoleByCode(ctx context.Context, code string) (*dto.RoleResponse, error) {
	role, err := s.roleRepo.GetByCode(ctx, code)
	if err != nil {
		return nil, err
	}

	return s.convertToRoleResponse(role), nil
}

// ListRoles 获取角色列表
func (s *roleService) ListRoles(ctx context.Context, req *dto.RoleListRequest) (*dto.RoleListResponse, error) {
	// 默认值处理
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Size <= 0 {
		req.Size = 10
	}

	// 获取角色列表
	roles, total, err := s.roleRepo.List(ctx, req.Name, req.Status, req.Page, req.Size)
	if err != nil {
		return nil, err
	}

	// 转换为响应格式
	result := &dto.RoleListResponse{
		Total: total,
		List:  make([]dto.RoleResponse, 0, len(roles)),
	}

	for _, role := range roles {
		result.List = append(result.List, *s.convertToRoleResponse(role))
	}

	return result, nil
}

// InitSystemRoles 初始化系统角色
func (s *roleService) InitSystemRoles(ctx context.Context) error {
	return s.roleRepo.InitSystemRoles(ctx)
}

// convertToRoleResponse 将角色实体转换为角色响应
func (s *roleService) convertToRoleResponse(role *entity.Role) *dto.RoleResponse {
	return &dto.RoleResponse{
		ID:          role.ID,
		Name:        role.Name,
		Description: role.Description,
		Code:        role.Code,
		Status:      role.Status,
		IsSystem:    role.IsSystem,
		CreatedAt:   role.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}

// 错误定义
var (
	ErrRoleCodeExists         = NewServiceError("角色编码已存在")
	ErrSystemRoleCannotDelete = NewServiceError("系统角色不允许删除")
)

// ServiceError 服务错误
type ServiceError struct {
	Message string
}

// Error 实现 error 接口
func (e *ServiceError) Error() string {
	return e.Message
}

// NewServiceError 创建服务错误
func NewServiceError(message string) *ServiceError {
	return &ServiceError{
		Message: message,
	}
}
