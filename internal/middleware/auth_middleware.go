package middleware

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/casbin/casbin/v2"
	"github.com/gin-gonic/gin"

	"oss-backend/internal/repository"
	"oss-backend/internal/service"
	"oss-backend/internal/utils"
	"oss-backend/pkg/common"
)

// 权限校验等级
const (
	LevelSystem  = "system"  // 系统级权限
	LevelGroup   = "group"   // 群组级权限
	LevelProject = "project" // 项目级权限
)

// AuthMiddleware 统一的认证与授权中间件
type AuthMiddleware struct {
	authService service.AuthService
	userRepo    repository.UserRepository
	enforcer    *casbin.Enforcer
}

// NewAuthMiddleware 创建认证与授权中间件
func NewAuthMiddleware(authService service.AuthService, userRepo repository.UserRepository, enforcer *casbin.Enforcer) *AuthMiddleware {
	return &AuthMiddleware{
		authService: authService,
		userRepo:    userRepo,
		enforcer:    enforcer,
	}
}

// AuthCheck 统一权限验证中间件
// level: 权限级别 (system/group/project)
// resourceType: 资源类型，用于从URL提取，如不指定则从URL路径提取
func (m *AuthMiddleware) AuthCheck(level string, resourceType string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从上下文获取用户ID
		userIDValue, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, common.ErrorResponse("未授权"))
			c.Abort()
			return
		}

		// 根据不同权限级别执行相应验证逻辑
		switch level {
		case LevelSystem:
			m.checkSystemPermission(c, userIDValue)
		case LevelGroup:
			m.checkGroupPermission(c, userIDValue, resourceType)
		case LevelProject:
			m.checkProjectPermission(c, userIDValue, resourceType)
		default:
			// 默认使用系统级权限检查
			m.checkSystemPermission(c, userIDValue)
		}
	}
}

// 系统级权限检查
func (m *AuthMiddleware) checkSystemPermission(c *gin.Context, userIDValue interface{}) {
	// 获取用户所有角色
	userRoles, exists := c.Get("userRoles")
	if !exists {
		c.JSON(http.StatusUnauthorized, common.ErrorResponse("用户角色信息不存在"))
		c.Abort()
		return
	}

	// 构造请求对象
	obj := c.Request.URL.Path
	act := c.Request.Method

	// 系统级域为 "system"
	domain := "system"

	// 检查每个角色的权限
	roles := userRoles.([]string)
	allowed := false

	for _, role := range roles {
		// 检查权限
		result, err := m.authService.CheckPermission(role, domain, obj, act)
		if err == nil && result {
			allowed = true
			break
		}
	}

	// 如果所有角色都没有权限，则拒绝访问
	if !allowed {
		c.JSON(http.StatusForbidden, common.ErrorResponse("权限不足"))
		c.Abort()
		return
	}

	c.Next()
}

// 群组级权限检查
func (m *AuthMiddleware) checkGroupPermission(c *gin.Context, userIDValue interface{}, resourceType string) {
	userID := userIDValue.(uint64)

	// 获取群组ID
	var groupID uint64

	// 首先尝试从路径参数获取
	groupIDStr := c.Param("id")
	if groupIDStr != "" {
		id, err := strconv.ParseUint(groupIDStr, 10, 64)
		if err == nil {
			groupID = id
		}
	}

	// 如果路径没有，尝试从查询参数获取
	if groupID == 0 {
		groupIDStr = c.Query("group_id")
		if groupIDStr != "" {
			id, err := strconv.ParseUint(groupIDStr, 10, 64)
			if err == nil {
				groupID = id
			}
		}
	}

	// 如果没有群组ID，则跳过权限检查
	if groupID == 0 {
		c.Next()
		return
	}

	// 构造域标识
	domain := fmt.Sprintf("group:%d", groupID)

	// 构造资源和操作
	var resource string
	if resourceType != "" {
		resource = resourceType
	} else {
		// 从URL路径中提取资源
		pathParts := strings.Split(c.Request.URL.Path, "/")
		if len(pathParts) > 4 {
			resource = pathParts[4]
		} else {
			resource = "*"
		}
	}

	// 从HTTP方法映射到操作
	act := utils.MapMethodToAction(c.Request.Method)

	// 构造用户标识
	sub := fmt.Sprintf("user:%d", userID)

	// 检查用户直接权限
	allowed, err := m.authService.CheckPermission(sub, domain, resource, act)
	if err == nil && allowed {
		c.Next()
		return
	}

	// 获取用户在此域的角色
	roles, err := m.authService.GetRolesForUser(sub, domain)
	if err != nil {
		c.JSON(http.StatusInternalServerError, common.ErrorResponse("获取用户角色失败"))
		c.Abort()
		return
	}

	// 检查角色权限
	for _, role := range roles {
		allowed, err := m.authService.CheckPermission(role, domain, resource, act)
		if err == nil && allowed {
			c.Next()
			return
		}
	}

	// 如果都没有权限，则拒绝访问
	c.JSON(http.StatusForbidden, common.ErrorResponse("您没有权限操作此群组资源"))
	c.Abort()
}

// 项目级权限检查
func (m *AuthMiddleware) checkProjectPermission(c *gin.Context, userIDValue interface{}, resourceType string) {
	userID := userIDValue.(uint64)

	// 获取项目ID
	var projectID uint64

	// 首先尝试从路径参数获取
	projectIDStr := c.Param("id")
	if projectIDStr != "" {
		id, err := strconv.ParseUint(projectIDStr, 10, 64)
		if err == nil {
			projectID = id
		}
	}

	// 如果路径没有，尝试从查询参数获取
	if projectID == 0 {
		projectIDStr = c.Query("project_id")
		if projectIDStr != "" {
			id, err := strconv.ParseUint(projectIDStr, 10, 64)
			if err == nil {
				projectID = id
			}
		}
	}

	// 如果没有项目ID，则跳过权限检查
	if projectID == 0 {
		c.Next()
		return
	}

	// 构造域标识
	domain := fmt.Sprintf("project:%d", projectID)

	// 构造资源和操作
	var resource string
	if resourceType != "" {
		resource = resourceType
	} else {
		// 从URL路径中提取资源
		pathParts := strings.Split(c.Request.URL.Path, "/")
		if len(pathParts) > 4 {
			resource = pathParts[4]
		} else {
			resource = "*"
		}
	}

	// 从HTTP方法映射到操作
	act := utils.MapMethodToAction(c.Request.Method)

	// 构造用户标识
	sub := fmt.Sprintf("user:%d", userID)

	// 检查用户直接权限
	allowed, err := m.authService.CheckPermission(sub, domain, resource, act)
	if err == nil && allowed {
		c.Next()
		return
	}

	// 获取用户在此域的角色
	roles, err := m.authService.GetRolesForUser(sub, domain)
	if err != nil {
		c.JSON(http.StatusInternalServerError, common.ErrorResponse("获取用户角色失败"))
		c.Abort()
		return
	}

	// 检查角色权限
	for _, role := range roles {
		allowed, err := m.authService.CheckPermission(role, domain, resource, act)
		if err == nil && allowed {
			c.Next()
			return
		}
	}

	// 如果都没有权限，则拒绝访问
	c.JSON(http.StatusForbidden, common.ErrorResponse("您没有权限操作此项目资源"))
	c.Abort()
}

// RequireRole 检查用户是否拥有特定角色
func (m *AuthMiddleware) RequireRole(roleName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从上下文中获取用户ID
		userIDValue, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, common.ErrorResponse("请先登录"))
			c.Abort()
			return
		}

		userID := userIDValue.(uint64)

		// 获取用户角色
		roles, err := m.userRepo.GetUserRoles(c, userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, common.ErrorResponse("获取用户角色失败"))
			c.Abort()
			return
		}

		// 检查用户是否拥有指定角色
		hasRole := false
		for _, role := range roles {
			if role.Code == roleName {
				hasRole = true
				break
			}
		}

		if !hasRole {
			c.JSON(http.StatusForbidden, common.ErrorResponse("权限不足:需要 "+roleName+" 角色"))
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAnyRole 要求用户具有任意指定角色之一
func (m *AuthMiddleware) RequireAnyRole(roleNames ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从上下文中获取用户ID
		userIDValue, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, common.ErrorResponse("请先登录"))
			c.Abort()
			return
		}

		userID := userIDValue.(uint64)

		// 获取用户角色
		roles, err := m.userRepo.GetUserRoles(c, userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, common.ErrorResponse("获取用户角色失败"))
			c.Abort()
			return
		}

		// 检查用户是否拥有任意指定角色
		hasRole := false
		for _, role := range roles {
			for _, roleName := range roleNames {
				if role.Code == roleName {
					hasRole = true
					break
				}
			}
			if hasRole {
				break
			}
		}

		if !hasRole {
			c.JSON(http.StatusForbidden, common.ErrorResponse("权限不足:需要管理员权限"))
			c.Abort()
			return
		}

		c.Next()
	}
}

// InitializeRBAC 初始化RBAC策略
func (m *AuthMiddleware) InitializeRBAC() error {
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
	_, err := m.enforcer.RemoveFilteredPolicy(0, "GROUP_ADMIN", "MEMBER")
	if err != nil {
		return err
	}

	// 添加新策略
	_, err = m.enforcer.AddPolicies(policies)
	return err
}

// RequireAdmin 检查用户是否是管理员
func (m *AuthMiddleware) RequireAdmin() gin.HandlerFunc {
	return m.RequireRole("GROUP_ADMIN")
}

// RequireProjectAdmin 检查用户是否是项目管理员（现在使用群组管理员代替）
func (m *AuthMiddleware) RequireProjectAdmin() gin.HandlerFunc {
	return m.RequireRole("GROUP_ADMIN")
}

// GetGroupIDFromParam 从URL参数中获取群组ID
func GetGroupIDFromParam(c *gin.Context) (string, error) {
	groupID := c.Param("groupID")
	if groupID == "" {
		return "0", nil
	}
	return groupID, nil
}

// GetGroupIDFromProjectParam 从项目参数中获取关联的群组ID
func GetGroupIDFromProjectParam(c *gin.Context) (string, error) {
	// 从路径参数中获取项目ID
	projectID := c.Param("projectID")
	if projectID == "" {
		return "0", nil
	}
	return "0", nil
}

// Authorize 基于资源和操作的授权中间件
// obj: 资源对象类型，如 "projects", "files"
// act: 操作类型，如 "create", "read"
// getDomainIDFunc: 获取域ID的函数（如群组ID或项目ID）
func (m *AuthMiddleware) Authorize(obj string, act string, getDomainIDFunc func(c *gin.Context) (string, error)) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从上下文获取用户ID
		userIDValue, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, common.ErrorResponse("未授权"))
			c.Abort()
			return
		}
		userID := userIDValue.(uint64)

		// 获取域标识
		domainID, err := getDomainIDFunc(c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, common.ErrorResponse("获取权限域失败"))
			c.Abort()
			return
		}

		// 如果域ID为空或"0"，表示全局资源，无需检查特定权限
		if domainID == "" || domainID == "0" {
			c.Next()
			return
		}

		// 构造用户标识
		sub := fmt.Sprintf("user:%d", userID)

		// 检查用户直接权限
		allowed, err := m.authService.CheckPermission(sub, domainID, obj, act)
		if err == nil && allowed {
			c.Next()
			return
		}

		// 获取用户在此域的角色
		roles, err := m.authService.GetRolesForUser(sub, domainID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, common.ErrorResponse("获取用户角色失败"))
			c.Abort()
			return
		}

		// 检查角色权限
		for _, role := range roles {
			allowed, err := m.authService.CheckPermission(role, domainID, obj, act)
			if err == nil && allowed {
				c.Next()
				return
			}
		}

		// 如果都没有权限，则拒绝访问
		c.JSON(http.StatusForbidden, common.ErrorResponse(fmt.Sprintf("您没有权限执行此操作: %s %s", act, obj)))
		c.Abort()
	}
}
