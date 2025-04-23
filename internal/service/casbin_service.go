package service

import (
	"errors"
	"fmt"
	"path/filepath"
	"sync"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"gorm.io/gorm"

	"oss-backend/internal/model/entity"
	"oss-backend/internal/repository"
)

// CasbinService 权限服务接口
type CasbinService interface {
	// 权限检查
	CheckPermission(sub string, domain string, obj string, act string) (bool, error)

	// 权限管理
	AddPolicy(sub string, domain string, obj string, act string) (bool, error)
	AddRoleForUser(user string, role string, domain string) (bool, error)
	RemovePolicy(sub string, domain string, obj string, act string) (bool, error)
	RemoveRoleForUser(user string, role string, domain string) (bool, error)
	GetRolesForUser(user string, domain string) ([]string, error)
	GetPermissionsForUser(user string, domain string) ([][]string, error)

	// 批量操作
	AddPolicies(rules [][]string) (bool, error)
	RemovePolicies(rules [][]string) (bool, error)

	// 初始化权限
	InitPermissions() error
}

// casbinService Casbin服务实现
type casbinService struct {
	enforcer      *casbin.SyncedEnforcer
	adapter       *gormadapter.Adapter
	userRepo      repository.UserRepository
	roleRepo      repository.RoleRepository
	groupRepo     repository.GroupRepository
	modelPath     string
	enforceLocker sync.RWMutex
}

// NewCasbinService 创建Casbin服务
func NewCasbinService(db *gorm.DB, userRepo repository.UserRepository, roleRepo repository.RoleRepository, groupRepo repository.GroupRepository) (CasbinService, error) {
	// 初始化Casbin表
	if err := db.AutoMigrate(&entity.CasbinRule{}); err != nil {
		return nil, err
	}

	// 初始化适配器
	adapter, err := gormadapter.NewAdapterByDB(db)
	if err != nil {
		return nil, err
	}

	// 加载模型
	modelPath := filepath.Join("configs", "rbac_model.conf")

	// 创建增强器
	enforcer, err := casbin.NewSyncedEnforcer(modelPath, adapter)
	if err != nil {
		return nil, err
	}

	// 加载策略
	err = enforcer.LoadPolicy()
	if err != nil {
		return nil, err
	}

	service := &casbinService{
		enforcer:  enforcer,
		adapter:   adapter,
		userRepo:  userRepo,
		roleRepo:  roleRepo,
		groupRepo: groupRepo,
		modelPath: modelPath,
	}

	return service, nil
}

// NewCasbinServiceWithEnforcer 使用自定义配置创建Casbin服务(用于测试)
func NewCasbinServiceWithEnforcer(m model.Model, adapter *gormadapter.Adapter) (CasbinService, error) {
	enforcer, err := casbin.NewSyncedEnforcer(m, adapter)
	if err != nil {
		return nil, err
	}

	return &casbinService{
		enforcer: enforcer,
		adapter:  adapter,
	}, nil
}

// CheckPermission 检查权限
func (s *casbinService) CheckPermission(sub string, domain string, obj string, act string) (bool, error) {
	s.enforceLocker.RLock()
	defer s.enforceLocker.RUnlock()

	return s.enforcer.Enforce(sub, domain, obj, act)
}

// AddPolicy 添加策略
func (s *casbinService) AddPolicy(sub string, domain string, obj string, act string) (bool, error) {
	s.enforceLocker.Lock()
	defer s.enforceLocker.Unlock()

	return s.enforcer.AddPolicy(sub, domain, obj, act)
}

// AddRoleForUser 为用户添加角色
func (s *casbinService) AddRoleForUser(user string, role string, domain string) (bool, error) {
	s.enforceLocker.Lock()
	defer s.enforceLocker.Unlock()

	return s.enforcer.AddRoleForUserInDomain(user, role, domain)
}

// RemovePolicy 移除策略
func (s *casbinService) RemovePolicy(sub string, domain string, obj string, act string) (bool, error) {
	s.enforceLocker.Lock()
	defer s.enforceLocker.Unlock()

	return s.enforcer.RemovePolicy(sub, domain, obj, act)
}

// RemoveRoleForUser 移除用户角色
func (s *casbinService) RemoveRoleForUser(user string, role string, domain string) (bool, error) {
	s.enforceLocker.Lock()
	defer s.enforceLocker.Unlock()

	return s.enforcer.DeleteRoleForUserInDomain(user, role, domain)
}

// GetRolesForUser 获取用户角色
func (s *casbinService) GetRolesForUser(user string, domain string) ([]string, error) {
	s.enforceLocker.RLock()
	defer s.enforceLocker.RUnlock()

	roles := s.enforcer.GetRolesForUserInDomain(user, domain)
	return roles, nil
}

// GetPermissionsForUser 获取用户权限
func (s *casbinService) GetPermissionsForUser(user string, domain string) ([][]string, error) {
	s.enforceLocker.RLock()
	defer s.enforceLocker.RUnlock()

	permissions := s.enforcer.GetPermissionsForUserInDomain(user, domain)
	return permissions, nil
}

// AddPolicies 批量添加策略
func (s *casbinService) AddPolicies(rules [][]string) (bool, error) {
	s.enforceLocker.Lock()
	defer s.enforceLocker.Unlock()

	return s.enforcer.AddPolicies(rules)
}

// RemovePolicies 批量移除策略
func (s *casbinService) RemovePolicies(rules [][]string) (bool, error) {
	s.enforceLocker.Lock()
	defer s.enforceLocker.Unlock()

	return s.enforcer.RemovePolicies(rules)
}

// InitPermissions 初始化权限
func (s *casbinService) InitPermissions() error {
	if s.userRepo == nil || s.roleRepo == nil || s.groupRepo == nil {
		return errors.New("repository不能为空")
	}

	// 清空所有策略
	s.enforceLocker.Lock()
	defer s.enforceLocker.Unlock()

	_, err := s.enforcer.RemoveFilteredPolicy(0, "", "", "", "")
	if err != nil {
		return err
	}

	// 1. 初始化系统级别角色权限
	systemPolicies := [][]string{
		// 超级管理员拥有系统级所有权限
		{"SUPER_ADMIN", "system", "*", "*"},

		// 群组管理员可以管理群组和项目
		{"GROUP_ADMIN", "system", "/api/oss/group/*", "*"},
		{"GROUP_ADMIN", "system", "/api/oss/project/*", "*"},

		// 项目管理员可以管理项目成员
		{"PROJECT_ADMIN", "system", "/api/oss/project/member/*", "*"},

		// 普通成员可以查看基本信息
		{"MEMBER", "system", "/api/oss/user/info", "GET"},
	}

	_, err = s.enforcer.AddPolicies(systemPolicies)
	if err != nil {
		return err
	}

	// 2. 为不同域配置默认权限 - 群组级别
	groupPolicies := [][]string{
		// 群组管理员角色具有所有群组资源权限
		{"admin", "group", "*", "*"},

		// 普通成员角色
		{"member", "group", "file", "read"},
		{"member", "group", "file", "upload"},
		{"member", "group", "file", "download"},
	}

	_, err = s.enforcer.AddPolicies(groupPolicies)
	if err != nil {
		return err
	}

	// 3. 为不同域配置默认权限 - 项目级别
	projectPolicies := [][]string{
		// 项目管理员角色
		{"admin", "project", "*", "*"},

		// 编辑者角色
		{"editor", "project", "file", "*"},
		{"editor", "project", "member", "read"},

		// 查看者角色
		{"viewer", "project", "file", "read"},
		{"viewer", "project", "file", "download"},
	}

	_, err = s.enforcer.AddPolicies(projectPolicies)
	if err != nil {
		return err
	}

	// 4. 加载系统中已存在的用户角色
	// 这里只是示例，实际应从数据库加载
	// 从用户-角色表中加载

	// 保存策略
	return s.enforcer.SavePolicy()
}

// GetUserPermissions 获取特定用户在指定域的所有权限
func (s *casbinService) GetUserPermissions(userID uint64, domainType string, domainID string) ([]map[string]string, error) {
	// 构造用户标识
	sub := fmt.Sprintf("user:%d", userID)
	domain := fmt.Sprintf("%s:%s", domainType, domainID)

	// 获取用户在此域的所有角色
	roles, err := s.GetRolesForUser(sub, domain)
	if err != nil {
		return nil, err
	}

	// 获取用户直接权限
	userPerms, err := s.GetPermissionsForUser(sub, domain)
	if err != nil {
		return nil, err
	}

	// 获取角色权限
	var allPerms []map[string]string

	// 添加用户直接权限
	for _, perm := range userPerms {
		if len(perm) >= 4 {
			allPerms = append(allPerms, map[string]string{
				"subject": perm[0],
				"domain":  perm[1],
				"object":  perm[2],
				"action":  perm[3],
			})
		}
	}

	// 添加角色权限
	for _, role := range roles {
		rolePerms, err := s.GetPermissionsForUser(role, domain)
		if err != nil {
			continue
		}

		for _, perm := range rolePerms {
			if len(perm) >= 4 {
				allPerms = append(allPerms, map[string]string{
					"subject": perm[0],
					"domain":  perm[1],
					"object":  perm[2],
					"action":  perm[3],
				})
			}
		}
	}

	return allPerms, nil
}
