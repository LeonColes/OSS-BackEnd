package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"oss-backend/internal/repository"
)

// RoleAuthMiddleware 角色认证中间件
type RoleAuthMiddleware struct {
	userRepo repository.UserRepository
}

// NewRoleAuthMiddleware 创建角色认证中间件
func NewRoleAuthMiddleware(userRepo repository.UserRepository) *RoleAuthMiddleware {
	return &RoleAuthMiddleware{
		userRepo: userRepo,
	}
}

// RequireRole 要求用户具有特定角色
func (m *RoleAuthMiddleware) RequireRole(roleName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从上下文中获取用户ID
		userIDValue, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "未授权:需要登录",
			})
			c.Abort()
			return
		}

		userID := uint64(userIDValue.(uint))

		// 获取用户角色
		roles, err := m.userRepo.GetUserRoles(c, userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "服务器错误:获取用户角色失败",
			})
			c.Abort()
			return
		}

		// 检查用户是否拥有指定角色
		hasRole := false
		for _, role := range roles {
			if role.Name == roleName {
				hasRole = true
				break
			}
		}

		if !hasRole {
			c.JSON(http.StatusForbidden, gin.H{
				"code":    403,
				"message": "权限不足:需要 " + roleName + " 角色",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAnyRole 要求用户具有任意指定角色之一
func (m *RoleAuthMiddleware) RequireAnyRole(roleNames ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从上下文中获取用户ID
		userIDValue, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "未授权:需要登录",
			})
			c.Abort()
			return
		}

		userID := uint64(userIDValue.(uint))

		// 获取用户角色
		roles, err := m.userRepo.GetUserRoles(c, userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "服务器错误:获取用户角色失败",
			})
			c.Abort()
			return
		}

		// 检查用户是否拥有任意指定角色
		hasRole := false
		for _, role := range roles {
			for _, roleName := range roleNames {
				if role.Name == roleName {
					hasRole = true
					break
				}
			}
			if hasRole {
				break
			}
		}

		if !hasRole {
			c.JSON(http.StatusForbidden, gin.H{
				"code":    403,
				"message": "权限不足:需要管理员权限",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
