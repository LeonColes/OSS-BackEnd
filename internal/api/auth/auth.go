package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/leoncoles/oss-backend/internal/service"
)

// Controller 认证控制器
type Controller struct {
	authService service.AuthService
}

// NewController 创建认证控制器
func NewController(authService service.AuthService) *Controller {
	return &Controller{
		authService: authService,
	}
}

// Register 用户注册
func (c *Controller) Register(ctx *gin.Context) {
	var req service.RegisterRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的请求参数: " + err.Error(),
		})
		return
	}

	userID, err := c.authService.Register(ctx, &req)
	if err != nil {
		var statusCode int
		var errMsg string

		if err == service.ErrUserExists {
			statusCode = http.StatusConflict
			errMsg = "该邮箱已注册"
		} else {
			statusCode = http.StatusInternalServerError
			errMsg = "注册失败: " + err.Error()
		}

		ctx.JSON(statusCode, gin.H{
			"code":    statusCode,
			"message": errMsg,
		})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"code":    201,
		"message": "注册成功",
		"data": gin.H{
			"user_id": userID,
		},
	})
}

// Login 用户登录
func (c *Controller) Login(ctx *gin.Context) {
	var req service.LoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的请求参数: " + err.Error(),
		})
		return
	}

	// 获取客户端IP
	clientIP := ctx.ClientIP()

	// 执行登录
	token, user, err := c.authService.Login(ctx, &req, clientIP)
	if err != nil {
		var statusCode int
		var errMsg string

		if err == service.ErrInvalidCredentials {
			statusCode = http.StatusUnauthorized
			errMsg = "邮箱或密码不正确"
		} else {
			statusCode = http.StatusInternalServerError
			errMsg = "登录失败: " + err.Error()
		}

		ctx.JSON(statusCode, gin.H{
			"code":    statusCode,
			"message": errMsg,
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "登录成功",
		"data": gin.H{
			"token":     token,
			"user_info": user,
		},
	})
}

// RefreshToken 刷新访问令牌
func (c *Controller) RefreshToken(ctx *gin.Context) {
	// 获取请求参数
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的请求参数: " + err.Error(),
		})
		return
	}

	// 刷新令牌
	token, err := c.authService.RefreshToken(ctx, req.RefreshToken)
	if err != nil {
		var statusCode int
		var errMsg string

		if err == service.ErrInvalidToken {
			statusCode = http.StatusUnauthorized
			errMsg = "无效的刷新令牌"
		} else if err == service.ErrTokenExpired {
			statusCode = http.StatusUnauthorized
			errMsg = "刷新令牌已过期，请重新登录"
		} else {
			statusCode = http.StatusInternalServerError
			errMsg = "刷新令牌失败: " + err.Error()
		}

		ctx.JSON(statusCode, gin.H{
			"code":    statusCode,
			"message": errMsg,
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "刷新令牌成功",
		"data":    token,
	})
}

// GetUserInfo 获取当前用户信息
func (c *Controller) GetUserInfo(ctx *gin.Context) {
	// 从上下文获取用户ID
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "未授权：用户未登录",
		})
		return
	}

	// 获取用户信息
	user, err := c.authService.GetUserByID(ctx, uint64(userID.(uint)))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取用户信息失败: " + err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取用户信息成功",
		"data": gin.H{
			"user_info": user,
		},
	})
}

// ChangePassword 修改密码
func (c *Controller) ChangePassword(ctx *gin.Context) {
	// 获取请求参数
	var req struct {
		OldPassword string `json:"old_password" binding:"required"`
		NewPassword string `json:"new_password" binding:"required,min=6"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的请求参数: " + err.Error(),
		})
		return
	}

	// 从上下文获取用户ID
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "未授权：用户未登录",
		})
		return
	}

	// 修改密码
	err := c.authService.ChangePassword(ctx, uint64(userID.(uint)), req.OldPassword, req.NewPassword)
	if err != nil {
		var statusCode int
		var errMsg string

		if err == service.ErrInvalidCredentials {
			statusCode = http.StatusBadRequest
			errMsg = "原密码不正确"
		} else {
			statusCode = http.StatusInternalServerError
			errMsg = "修改密码失败: " + err.Error()
		}

		ctx.JSON(statusCode, gin.H{
			"code":    statusCode,
			"message": errMsg,
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "密码修改成功",
	})
}

// Logout 用户登出
func (c *Controller) Logout(ctx *gin.Context) {
	// 实际上服务器端不需要做任何操作，JWT令牌是无状态的
	// 客户端只需要删除本地存储的令牌即可
	// 但为了API的完整性，我们还是提供了这个接口

	ctx.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "登出成功",
	})
} 