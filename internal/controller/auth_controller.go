package controller

import (
	"net/http"

	"oss-backend/internal/service"

	"github.com/gin-gonic/gin"
)

// AuthController 处理认证相关请求
type AuthController struct {
	authService service.AuthService
	userService service.UserService
}

// NewAuthController 创建新的认证控制器
func NewAuthController(authService service.AuthService, userService service.UserService) *AuthController {
	return &AuthController{
		authService: authService,
		userService: userService,
	}
}

// Register 用户注册
// @Summary 用户注册
// @Description 注册新用户并返回用户ID
// @Tags auth
// @Accept json
// @Produce json
// @Param request body service.RegisterRequest true "注册信息"
// @Success 201 {object} map[string]interface{} "注册成功"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 409 {object} map[string]interface{} "用户已存在"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /auth/register [post]
func (c *AuthController) Register(ctx *gin.Context) {
	// TODO: 实现注册逻辑
	ctx.JSON(200, gin.H{"message": "注册功能待实现"})
}

// Login 用户登录
// @Summary 用户登录
// @Description 用户登录并获取认证令牌
// @Tags auth
// @Accept json
// @Produce json
// @Param request body service.LoginRequest true "登录信息"
// @Success 200 {object} map[string]interface{} "登录成功"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 401 {object} map[string]interface{} "邮箱或密码不正确"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /auth/login [post]
func (c *AuthController) Login(ctx *gin.Context) {
	// TODO: 实现登录逻辑
	ctx.JSON(200, gin.H{"message": "登录功能待实现"})
}

// RefreshToken 刷新访问令牌
// @Summary 刷新访问令牌
// @Description 使用刷新令牌获取新的访问令牌
// @Tags auth
// @Accept json
// @Produce json
// @Param request body map[string]string true "刷新令牌请求"
// @Success 200 {object} map[string]interface{} "刷新成功"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 401 {object} map[string]interface{} "无效的刷新令牌或已过期"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /auth/refresh [post]
func (c *AuthController) RefreshToken(ctx *gin.Context) {
	// TODO: 实现刷新令牌逻辑
	ctx.JSON(200, gin.H{"message": "刷新令牌功能待实现"})
}

// GetUserInfo 获取当前用户信息
// @Summary 获取用户信息
// @Description 获取当前登录用户的信息
// @Tags user
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "获取成功"
// @Failure 401 {object} map[string]interface{} "未授权：用户未登录"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /user/info [get]
func (c *AuthController) GetUserInfo(ctx *gin.Context) {
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
	user, err := c.userService.GetUserInfo(ctx, uint(userID.(uint)))
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
// @Summary 修改用户密码
// @Description 修改当前登录用户的密码
// @Tags user
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body map[string]string true "密码信息"
// @Success 200 {object} map[string]interface{} "修改成功"
// @Failure 400 {object} map[string]interface{} "请求参数错误或原密码不正确"
// @Failure 401 {object} map[string]interface{} "未授权：用户未登录"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /user/password [put]
func (c *AuthController) ChangePassword(ctx *gin.Context) {
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
	err := c.userService.UpdatePassword(ctx, uint(userID.(uint)), req)
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
// @Summary 用户登出
// @Description 用户登出，使当前令牌失效
// @Tags user
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "登出成功"
// @Router /user/logout [post]
func (c *AuthController) Logout(ctx *gin.Context) {
	// 实际上服务器端不需要做任何操作，JWT令牌是无状态的
	// 客户端只需要删除本地存储的令牌即可
	// 但为了API的完整性，我们还是提供了这个接口

	ctx.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "登出成功",
	})
}
