package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"

	"oss-backend/pkg/common"
)

// 使用与服务层相同的JWT密钥，实际应用中应从配置文件读取
var jwtSecret = []byte("oss-backend-secret-key")

// JWTClaims JWT声明
type JWTClaims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

// JWTAuthMiddleware JWT认证中间件
type JWTAuthMiddleware struct{}

// NewJWTAuthMiddleware 创建JWT认证中间件
func NewJWTAuthMiddleware() *JWTAuthMiddleware {
	return &JWTAuthMiddleware{}
}

// AuthMiddleware 认证中间件
func (m *JWTAuthMiddleware) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头获取token
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, common.ErrorResponse("未授权:请先登录"))
			c.Abort()
			return
		}

		// 处理Bearer前缀
		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			c.JSON(http.StatusUnauthorized, common.ErrorResponse("未授权:token格式错误"))
			c.Abort()
			return
		}

		// 解析token
		token, err := jwt.ParseWithClaims(parts[1], &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
			// 验证算法
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return jwtSecret, nil
		})

		if err != nil {
			c.JSON(http.StatusUnauthorized, common.ErrorResponse("未授权:token无效"))
			c.Abort()
			return
		}

		// 验证token
		if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
			// 检查token是否过期
			if time.Now().After(claims.ExpiresAt.Time) {
				c.JSON(http.StatusUnauthorized, common.ErrorResponse("未授权:token已过期"))
				c.Abort()
				return
			}

			// 设置用户ID到上下文
			c.Set("userID", claims.UserID)
			c.Set("userEmail", claims.Email)
			c.Next()
			return
		}

		c.JSON(http.StatusUnauthorized, common.ErrorResponse("未授权:token无效"))
		c.Abort()
	}
}
