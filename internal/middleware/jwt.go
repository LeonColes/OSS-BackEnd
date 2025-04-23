package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

// JWTAuthMiddleware JWT认证中间件
type JWTAuthMiddleware struct {
	Secret string
}

// NewJWTAuthMiddleware 创建JWT认证中间件
func NewJWTAuthMiddleware() *JWTAuthMiddleware {
	return &JWTAuthMiddleware{
		Secret: "your-secret-key", // 实际应用中应从配置中读取
	}
}

// AuthMiddleware 认证中间件
func (m *JWTAuthMiddleware) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头获取token
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "未授权:请先登录",
			})
			c.Abort()
			return
		}

		// 处理Bearer前缀
		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "未授权:token格式错误",
			})
			c.Abort()
			return
		}

		// 解析token
		token, err := jwt.Parse(parts[1], func(token *jwt.Token) (interface{}, error) {
			// 验证算法
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(m.Secret), nil
		})

		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "未授权:token无效",
			})
			c.Abort()
			return
		}

		// 验证token
		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			// 检查token是否过期
			if !m.verifyExpiration(claims) {
				c.JSON(http.StatusUnauthorized, gin.H{
					"code":    401,
					"message": "未授权:token已过期",
				})
				c.Abort()
				return
			}

			// 设置用户ID到上下文
			if userID, ok := claims["id"]; ok {
				c.Set("userID", uint(userID.(float64)))
				c.Next()
				return
			}
		}

		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "未授权:token无效",
		})
		c.Abort()
	}
}

// 验证token是否过期
func (m *JWTAuthMiddleware) verifyExpiration(claims jwt.MapClaims) bool {
	if exp, ok := claims["exp"]; ok {
		expTime := time.Unix(int64(exp.(float64)), 0)
		return expTime.After(time.Now())
	}
	return false
}
