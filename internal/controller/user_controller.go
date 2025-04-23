package controller

import (
	"net/http"
	"oss-backend/internal/service"

	"github.com/gin-gonic/gin"
)

// UserController 处理用户相关请求
type UserController struct {
	userService service.UserService
}

// NewUserController 创建新的用户控制器
func NewUserController(userService service.UserService) *UserController {
	return &UserController{
		userService: userService,
	}
}

// GetUserInfo 获取用户信息
func (c *UserController) GetUserInfo(ctx *gin.Context) {
	// 获取当前用户ID
	userID, exists := ctx.Get("userID")
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
		"data":    user,
	})
}

// UpdateUserInfo 更新用户信息
func (c *UserController) UpdateUserInfo(ctx *gin.Context) {
	var req struct {
		Nickname string `json:"nickname"`
		Avatar   string `json:"avatar"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的请求参数: " + err.Error(),
		})
		return
	}

	// 获取当前用户ID
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "未授权：用户未登录",
		})
		return
	}

	// 构造 UserRequest 并传递给 service 层
	userReq := service.UserRequest{
		Name:   req.Nickname,
		Avatar: req.Avatar,
	}

	// 更新用户信息
	user, err := c.userService.UpdateUserInfo(ctx, uint(userID.(uint)), userReq)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "更新用户信息失败: " + err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "用户信息更新成功",
		"data":    user,
	})
}

// UpdatePassword 更新密码
func (c *UserController) UpdatePassword(ctx *gin.Context) {
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

	// 获取当前用户ID
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "未授权：用户未登录",
		})
		return
	}

	// 更新密码
	err := c.userService.UpdatePassword(ctx, uint(userID.(uint)), req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "修改密码失败: " + err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "密码修改成功",
	})
}
