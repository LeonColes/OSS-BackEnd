package middleware

import (
	"fmt"
	"net/http"
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
	// 获取用户ID
	userID := userIDValue.(string)

	// 获取请求对象和操作
	obj := c.Request.URL.Path                        // 或者更精确的资源标识
	act := utils.MapMethodToAction(c.Request.Method) // 使用映射函数

	// 系统级域为 "system"
	const domain = "system"

	// 直接调用 CanUserAccessResource
	allowed, err := m.authService.CanUserAccessResource(c, userID, obj, act, domain)
	if err != nil {
		c.JSON(http.StatusInternalServerError, common.ErrorResponse("检查权限失败: "+err.Error()))
		c.Abort()
		return
	}

	// 如果没有权限，则拒绝访问
	if !allowed {
		c.JSON(http.StatusForbidden, common.ErrorResponse("权限不足"))
		c.Abort()
		return
	}

	c.Next()
}

// 群组级权限检查
func (m *AuthMiddleware) checkGroupPermission(c *gin.Context, userIDValue interface{}, resourceType string) {
	userID := userIDValue.(string)

	// 获取群组ID
	var groupID string

	// 首先尝试从路径参数获取
	groupIDStr := c.Param("id")
	if groupIDStr != "" {
		groupID = groupIDStr
	}

	// 如果路径没有，尝试从查询参数获取
	if groupID == "" {
		groupIDStr = c.Query("group_id")
		if groupIDStr != "" {
			groupID = groupIDStr
		}
	}

	// 如果没有群组ID，则跳过权限检查
	if groupID == "" {
		c.Next()
		return
	}

	// 构造域标识
	domain := fmt.Sprintf("group:%s", groupID)

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

	// 直接调用 CanUserAccessResource
	allowed, err := m.authService.CanUserAccessResource(c, userID, resource, act, domain)
	if err != nil {
		c.JSON(http.StatusInternalServerError, common.ErrorResponse("检查权限失败: "+err.Error()))
		c.Abort()
		return
	}

	// 如果没有权限，则拒绝访问
	if !allowed {
		c.JSON(http.StatusForbidden, common.ErrorResponse("您没有权限操作此群组资源"))
		c.Abort()
		return
	}

	c.Next()
}

// 项目级权限检查
func (m *AuthMiddleware) checkProjectPermission(c *gin.Context, userIDValue interface{}, resourceType string) {
	userID := userIDValue.(string)

	// 获取项目ID
	var projectID string

	// 首先尝试从路径参数获取
	projectIDStr := c.Param("id")
	if projectIDStr != "" {
		projectID = projectIDStr
	}

	// 如果路径没有，尝试从查询参数获取
	if projectID == "" {
		projectIDStr = c.Query("project_id")
		if projectIDStr != "" {
			projectID = projectIDStr
		}
	}

	// 如果没有项目ID，则跳过权限检查
	if projectID == "" {
		c.Next()
		return
	}

	// 构造域标识
	domain := fmt.Sprintf("project:%s", projectID)

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

	// 直接调用 CanUserAccessResource
	allowed, err := m.authService.CanUserAccessResource(c, userID, resource, act, domain)
	if err != nil {
		c.JSON(http.StatusInternalServerError, common.ErrorResponse("检查权限失败: "+err.Error()))
		c.Abort()
		return
	}

	// 如果没有权限，则拒绝访问
	if !allowed {
		c.JSON(http.StatusForbidden, common.ErrorResponse("您没有权限操作此项目资源"))
		c.Abort()
		return
	}

	c.Next()
}

// RequireRole 检查用户是否拥有特定角色 (假定在 "system" 域)
// 注意: 如果需要检查特定域的角色，需要修改此中间件以获取 domain
func (m *AuthMiddleware) RequireRole(roleName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从上下文中获取用户ID
		userIDValue, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, common.ErrorResponse("请先登录"))
			c.Abort()
			return
		}
		userID := userIDValue.(string)
		const domain = "system" // 假定检查系统域角色，如果需要检查其他域，需要动态获取

		// 使用 AuthService 检查角色
		hasRole, err := m.authService.IsUserInRole(c, userID, roleName, domain)
		if err != nil {
			c.JSON(http.StatusInternalServerError, common.ErrorResponse("检查角色失败: "+err.Error()))
			c.Abort()
			return
		}

		// 检查用户是否拥有指定角色
		// Remove old loop based on userRepo.GetUserRoles

		if !hasRole {
			c.JSON(http.StatusForbidden, common.ErrorResponse("权限不足:需要 "+roleName+" 角色"))
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAnyRole 要求用户具有任意指定角色之一 (假定在 "system" 域)
// 注意: 如果需要检查特定域的角色，需要修改此中间件以获取 domain
func (m *AuthMiddleware) RequireAnyRole(roleNames ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从上下文中获取用户ID
		userIDValue, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, common.ErrorResponse("请先登录"))
			c.Abort()
			return
		}
		userID := userIDValue.(string)
		const domain = "system" // 假定检查系统域角色，如果需要检查其他域，需要动态获取

		// 检查用户是否拥有任意指定角色
		hasRole := false
		for _, roleName := range roleNames {
			allowed, err := m.authService.IsUserInRole(c, userID, roleName, domain)
			if err != nil {
				// Consider logging the error but continuing the loop
				// For now, fail fast
				c.JSON(http.StatusInternalServerError, common.ErrorResponse("检查角色失败: "+err.Error()))
				c.Abort()
				return
			}
			if allowed {
				hasRole = true
				break
			}
		}

		// Remove old loop based on userRepo.GetUserRoles

		if !hasRole {
			// Improved error message clarity
			requiredRolesStr := strings.Join(roleNames, " 或 ")
			c.JSON(http.StatusForbidden, common.ErrorResponse("权限不足: 需要 "+requiredRolesStr+" 角色中的至少一个"))
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAdmin 检查用户是否是系统管理员 ("system" 域的 "ADMIN" 角色)
func (m *AuthMiddleware) RequireAdmin() gin.HandlerFunc {
	// Use the updated RequireRole which now checks 'system' domain by default
	return m.RequireRole("ADMIN")
}

// RequireProjectAdmin 检查用户是否是项目管理员
// 这需要从请求中获取 projectID 并检查特定角色 (例如 "PROJECT_ADMIN" 或 "admin") 在项目域 ("project:projectID")
func (m *AuthMiddleware) RequireProjectAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从上下文中获取用户ID
		userIDValue, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, common.ErrorResponse("请先登录"))
			c.Abort()
			return
		}
		userID := userIDValue.(string)

		// 从路径参数获取项目ID (需要根据实际路由调整参数名)
		projectID := c.Param("projectID") // 或者 c.Param("id") 等
		if projectID == "" {
			// 如果项目ID在其他地方 (e.g., query param, request body), 需要相应修改
			c.JSON(http.StatusBadRequest, common.ErrorResponse("请求路径缺少项目ID"))
			c.Abort()
			return
		}

		// 构造项目域
		projectDomain := fmt.Sprintf("project:%s", projectID)
		// 定义项目管理员角色代码 (需要与 Casbin 策略一致)
		const projectAdminRole = "PROJECT_ADMIN" // 或者使用 "admin" 如果策略是这样定义的

		// 检查用户是否在项目域拥有管理员角色
		hasRole, err := m.authService.IsUserInRole(c, userID, projectAdminRole, projectDomain)
		if err != nil {
			c.JSON(http.StatusInternalServerError, common.ErrorResponse("检查项目管理员权限失败: "+err.Error()))
			c.Abort()
			return
		}

		if !hasRole {
			c.JSON(http.StatusForbidden, common.ErrorResponse("权限不足: 需要项目管理员角色"))
			c.Abort()
			return
		}

		c.Next()
	}
}

// GetGroupIDFromParam 从URL参数中获取群组ID
func GetGroupIDFromParam(c *gin.Context) (string, error) {
	groupID := c.Param("groupID")
	if groupID == "" {
		return "", nil
	}
	return fmt.Sprintf("group:%s", groupID), nil
}

// GetDomainIDFromProjectParam 从项目参数中获取项目域ID
// 重命名并修复逻辑
func GetDomainIDFromProjectParam(c *gin.Context) (string, error) {
	// 尝试从路径参数获取项目ID (根据路由调整 "projectID" 或 "id")
	projectID := c.Param("projectID")
	if projectID == "" {
		projectID = c.Param("id")
	}
	// 如果路径没有，尝试从查询参数获取
	if projectID == "" {
		projectID = c.Query("project_id")
	}

	if projectID == "" {
		// 如果没有项目ID，返回空字符串表示无特定项目域
		return "", nil
	}
	// 返回构造好的项目域
	return fmt.Sprintf("project:%s", projectID), nil
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
		userID := userIDValue.(string)

		// 获取域标识
		domainID, err := getDomainIDFunc(c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, common.ErrorResponse("获取权限域失败"))
			c.Abort()
			return
		}

		// 如果域ID为空，表示全局资源或无法确定域，可能跳过检查或执行默认检查
		if domainID == "" {
			// 根据业务逻辑决定：是允许访问，还是拒绝，或是检查系统级权限？
			// 暂定为跳过检查 (允许访问)
			c.Next()
			return
		}

		// 直接调用 CanUserAccessResource
		allowed, err := m.authService.CanUserAccessResource(c, userID, obj, act, domainID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, common.ErrorResponse("检查权限失败: "+err.Error()))
			c.Abort()
			return
		}

		// 如果没有权限，则拒绝访问
		if !allowed {
			c.JSON(http.StatusForbidden, common.ErrorResponse(fmt.Sprintf("您没有权限执行此操作: %s %s on domain %s", act, obj, domainID)))
			c.Abort()
			return
		}

		c.Next()
	}
}
