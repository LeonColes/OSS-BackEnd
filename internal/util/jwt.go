package util

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/leoncoles/oss-backend/internal/config"
)

// 定义错误
var (
	ErrInvalidToken = errors.New("无效的令牌")
	ErrExpiredToken = errors.New("令牌已过期")
)

// UserClaims 用户JWT声明
type UserClaims struct {
	UserID uint64 `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

// GenerateToken 生成JWT令牌
func GenerateToken(userID uint64, email string, isRefresh bool) (string, error) {
	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		return "", err
	}

	// 设置过期时间
	expireTime := cfg.JWT.ExpireHours
	if isRefresh {
		expireTime = cfg.JWT.RefreshExpireHours
	}
	
	// 创建JWT声明
	claims := UserClaims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * time.Duration(expireTime))),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "oss-backend",
		},
	}

	// 创建令牌
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// 签名令牌
	return token.SignedString([]byte(cfg.JWT.Secret))
}

// ParseToken 解析验证JWT令牌
func ParseToken(tokenString string) (*UserClaims, error) {
	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	// 解析令牌
	token, err := jwt.ParseWithClaims(tokenString, &UserClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(cfg.JWT.Secret), nil
	})

	// 处理解析错误
	if err != nil {
		if ve, ok := err.(*jwt.ValidationError); ok {
			if ve.Errors&jwt.ValidationErrorExpired != 0 {
				return nil, ErrExpiredToken
			}
		}
		return nil, ErrInvalidToken
	}

	// 验证令牌
	if claims, ok := token.Claims.(*UserClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, ErrInvalidToken
} 