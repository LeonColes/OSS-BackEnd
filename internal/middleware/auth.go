package middleware

import (
	"net/http"
	"strings"

	"oss-backend/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"honnef.co/go/tools/config"
)

// JWTMiddleware JWT认证中间件结构体
type JWTMiddleware struct {
	authService service.AuthService
}

// NewJWTMiddleware 创建JWT中间件
func NewJWTMiddleware(authService service.AuthService) *JWTMiddleware {
	return &JWTMiddleware{
		authService: authService,
	}
}

// AuthMiddleware 认证中间件函数
func (m *JWTMiddleware) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头获取token
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未授权：缺少Authorization请求头"})
			c.Abort()
			return
		}

		// 解析Bearer令牌
		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未授权：授权格式错误"})
			c.Abort()
			return
		}

		// 获取令牌
		tokenString := parts[1]

		// 验证令牌
		claims, err := m.authService.ValidateToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未授权：" + err.Error()})
			c.Abort()
			return
		}

		// 将用户ID和角色存入上下文
		c.Set("userID", claims.UserID)
		c.Set("userEmail", claims.Email)

		c.Next()
	}
}

// JWTAuth JWT认证中间件
func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头获取token
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未授权：缺少Authorization请求头"})
			c.Abort()
			return
		}

		// 解析Bearer令牌
		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未授权：授权格式错误"})
			c.Abort()
			return
		}

		// 获取令牌
		tokenString := parts[1]

		// 解析令牌
		token, err := jwt.ParseWithClaims(tokenString, &jwtUserClaims{}, func(token *jwt.Token) (interface{}, error) {
			// 从配置获取JWT密钥
			cfg, err := config.Load()
			if err != nil {
				return nil, err
			}
			return []byte(cfg.JWT.Secret), nil
		})

		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未授权：" + err.Error()})
			c.Abort()
			return
		}

		// 验证令牌
		if claims, ok := token.Claims.(*jwtUserClaims); ok && token.Valid {
			// 将用户信息存入上下文
			c.Set("userID", claims.UserID)
			c.Set("email", claims.Email)
			c.Next()
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未授权：无效的令牌"})
			c.Abort()
			return
		}
	}
}

// jwtUserClaims JWT令牌声明结构体
type jwtUserClaims struct {
	UserID uint   `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

// RequireRoles 验证角色权限
func RequireRoles(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从上下文获取用户ID
		_, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未授权：用户未登录"})
			c.Abort()
			return
		}

		// 这里应该查询数据库获取用户角色
		// 简单示例，实际项目中需要数据库查询
		userRole := "member" // 假设用户角色

		// 检查用户角色是否在允许的角色列表中
		allowed := false
		for _, role := range roles {
			if role == userRole {
				allowed = true
				break
			}
		}

		if !allowed {
			c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "无权操作：权限不足"})
			c.Abort()
			return
		}

		c.Next()
	}
}
