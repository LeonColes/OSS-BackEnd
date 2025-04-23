package middleware

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

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

// UnifiedCasbinMiddleware 统一的RBAC权限中间件
type UnifiedCasbinMiddleware struct {
	casbinService service.CasbinService
}

// NewUnifiedCasbinMiddleware 创建统一Casbin中间件
func NewUnifiedCasbinMiddleware(casbinService service.CasbinService) *UnifiedCasbinMiddleware {
	return &UnifiedCasbinMiddleware{
		casbinService: casbinService,
	}
}

// AuthCheck 统一权限验证中间件
// level: 权限级别 (system/group/project)
// resourceType: 资源类型，用于从URL提取，如不指定则从URL路径提取
func (m *UnifiedCasbinMiddleware) AuthCheck(level string, resourceType string) gin.HandlerFunc {
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
func (m *UnifiedCasbinMiddleware) checkSystemPermission(c *gin.Context, userIDValue interface{}) {
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
		result, err := m.casbinService.CheckPermission(role, domain, obj, act)
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
func (m *UnifiedCasbinMiddleware) checkGroupPermission(c *gin.Context, userIDValue interface{}, resourceType string) {
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

	// 如果查询参数也没有，尝试从JSON获取
	if groupID == 0 {
		// 这部分在实际应用中需要根据具体情况实现
		// 例如在请求头或表单中获取
	}

	// 如果没有群组ID，则跳过权限检查
	if groupID == 0 {
		c.Next()
		return
	}

	// 构造域标识
	domain := fmt.Sprintf("group:%d", groupID)

	// 构造资源和操作
	// 获取资源类型
	var resource string
	if resourceType != "" {
		resource = resourceType
	} else {
		// 从URL路径中提取资源
		pathParts := strings.Split(c.Request.URL.Path, "/")
		// 从路径中提取资源类型，例如/api/oss/group/file/xxx中的file
		if len(pathParts) > 4 {
			resource = pathParts[4] // 根据API路径结构调整索引
		} else {
			resource = "*"
		}
	}

	// 从HTTP方法映射到操作
	act := utils.MapMethodToAction(c.Request.Method)

	// 构造用户标识
	sub := fmt.Sprintf("user:%d", userID)

	// 检查用户直接权限
	allowed, err := m.casbinService.CheckPermission(sub, domain, resource, act)
	if err == nil && allowed {
		c.Next()
		return
	}

	// 获取用户在此域的角色
	roles, err := m.casbinService.GetRolesForUser(sub, domain)
	if err != nil {
		c.JSON(http.StatusInternalServerError, common.ErrorResponse("获取用户角色失败"))
		c.Abort()
		return
	}

	// 检查角色权限
	for _, role := range roles {
		allowed, err := m.casbinService.CheckPermission(role, domain, resource, act)
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
func (m *UnifiedCasbinMiddleware) checkProjectPermission(c *gin.Context, userIDValue interface{}, resourceType string) {
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

	// 如果查询参数也没有，尝试从JSON获取
	if projectID == 0 {
		// 这部分在实际应用中需要根据具体情况实现
	}

	// 如果没有项目ID，则跳过权限检查
	if projectID == 0 {
		c.Next()
		return
	}

	// 构造域标识
	domain := fmt.Sprintf("project:%d", projectID)

	// 构造资源和操作
	// 获取资源类型
	var resource string
	if resourceType != "" {
		resource = resourceType
	} else {
		// 从URL路径中提取资源
		pathParts := strings.Split(c.Request.URL.Path, "/")
		// 从路径中提取资源类型，例如/api/oss/project/file/xxx中的file
		if len(pathParts) > 4 {
			resource = pathParts[4] // 根据API路径结构调整索引
		} else {
			resource = "*"
		}
	}

	// 从HTTP方法映射到操作
	act := utils.MapMethodToAction(c.Request.Method)

	// 构造用户标识
	sub := fmt.Sprintf("user:%d", userID)

	// 检查用户直接权限
	allowed, err := m.casbinService.CheckPermission(sub, domain, resource, act)
	if err == nil && allowed {
		c.Next()
		return
	}

	// 获取用户在此域的角色
	roles, err := m.casbinService.GetRolesForUser(sub, domain)
	if err != nil {
		c.JSON(http.StatusInternalServerError, common.ErrorResponse("获取用户角色失败"))
		c.Abort()
		return
	}

	// 检查角色权限
	for _, role := range roles {
		allowed, err := m.casbinService.CheckPermission(role, domain, resource, act)
		if err == nil && allowed {
			c.Next()
			return
		}
	}

	// 如果都没有权限，则拒绝访问
	c.JSON(http.StatusForbidden, common.ErrorResponse("您没有权限操作此项目资源"))
	c.Abort()
}
