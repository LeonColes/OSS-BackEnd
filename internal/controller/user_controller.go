package controller

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"oss-backend/internal/model/dto"
	"oss-backend/internal/service"
	"oss-backend/pkg/common"
)

// UserController 用户控制器
type UserController struct {
	userService service.UserService
}

// NewUserController 创建用户控制器
func NewUserController(userService service.UserService) *UserController {
	return &UserController{
		userService: userService,
	}
}

// Register 用户注册
// @Summary 用户注册
// @Description 注册新用户
// @Tags 用户模块
// @Accept json
// @Produce json
// @Param request body dto.UserRegisterRequest true "注册信息"
// @Success 200 {object} common.Response "成功"
// @Failure 400 {object} common.Response "请求参数错误"
// @Failure 500 {object} common.Response "内部服务器错误"
// @Router /api/oss/user/register [post]
func (c *UserController) Register(ctx *gin.Context) {
	var req dto.UserRegisterRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, common.ErrorResponse(err.Error()))
		return
	}

	user, err := c.userService.Register(ctx, &req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, common.ErrorResponse(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, common.SuccessResponse(user))
}

// Login 用户登录
// @Summary 用户登录
// @Description 用户登录并获取令牌
// @Tags 用户模块
// @Accept json
// @Produce json
// @Param request body dto.UserLoginRequest true "登录信息"
// @Success 200 {object} common.Response{data=dto.LoginResponse} "成功"
// @Failure 400 {object} common.Response "请求参数错误"
// @Failure 500 {object} common.Response "内部服务器错误"
// @Router /api/oss/user/login [post]
func (c *UserController) Login(ctx *gin.Context) {
	var req dto.UserLoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, common.ErrorResponse(err.Error()))
		return
	}

	// 获取客户端IP
	clientIP := ctx.ClientIP()

	result, err := c.userService.Login(ctx, &req, clientIP)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, common.ErrorResponse(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, common.SuccessResponse(result))
}

// GetUserInfo 获取用户信息
// @Summary 获取用户信息
// @Description 获取当前登录用户的详细信息
// @Tags 用户模块
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer {{token}}"
// @Success 200 {object} common.Response{data=dto.UserResponse} "成功"
// @Failure 401 {object} common.Response "未授权"
// @Failure 500 {object} common.Response "内部服务器错误"
// @Router /api/oss/user/info [get]
func (c *UserController) GetUserInfo(ctx *gin.Context) {
	// 从上下文中获取用户ID
	userIDValue, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, common.ErrorResponse("未授权"))
		return
	}
	userID := userIDValue.(uint64)

	userInfo, err := c.userService.GetUserInfo(ctx, userID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, common.ErrorResponse(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, common.SuccessResponse(userInfo))
}

// UpdateUserInfo 更新用户信息
// @Summary 更新用户信息
// @Description 更新当前用户的基本信息
// @Tags 用户模块
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer {{token}}"
// @Param request body dto.UserUpdateRequest true "用户信息"
// @Success 200 {object} common.Response "成功"
// @Failure 400 {object} common.Response "请求参数错误"
// @Failure 401 {object} common.Response "未授权"
// @Failure 500 {object} common.Response "内部服务器错误"
// @Router /api/oss/user/update [post]
func (c *UserController) UpdateUserInfo(ctx *gin.Context) {
	var req dto.UserUpdateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, common.ErrorResponse(err.Error()))
		return
	}

	// 从上下文中获取用户ID
	userIDValue, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, common.ErrorResponse("未授权"))
		return
	}
	userID := userIDValue.(uint64)

	err := c.userService.UpdateUserInfo(ctx, userID, &req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, common.ErrorResponse(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, common.SuccessResponse(nil))
}

// UpdatePassword 更新密码
// @Summary 更新密码
// @Description 更新当前用户的密码
// @Tags 用户模块
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer {{token}}"
// @Param request body dto.UserPasswordUpdateRequest true "密码信息"
// @Success 200 {object} common.Response "成功"
// @Failure 400 {object} common.Response "请求参数错误"
// @Failure 401 {object} common.Response "未授权"
// @Failure 500 {object} common.Response "内部服务器错误"
// @Router /api/oss/user/password [post]
func (c *UserController) UpdatePassword(ctx *gin.Context) {
	var req dto.UserPasswordUpdateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, common.ErrorResponse(err.Error()))
		return
	}

	// 从上下文中获取用户ID
	userIDValue, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, common.ErrorResponse("未授权"))
		return
	}
	userID := userIDValue.(uint64)

	err := c.userService.UpdatePassword(ctx, userID, &req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, common.ErrorResponse(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, common.SuccessResponse(nil))
}

// ListUsers 获取用户列表
// @Summary 获取用户列表
// @Description 根据条件获取用户列表
// @Tags 用户模块
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer {{token}}"
// @Param email query string false "用户邮箱，模糊查询"
// @Param name query string false "用户姓名，模糊查询"
// @Param status query int false "状态：1-正常，2-禁用，3-锁定"
// @Param page query int false "页码，默认1"
// @Param size query int false "每页数量，默认10"
// @Success 200 {object} common.Response{data=dto.UserListResponse} "成功"
// @Failure 401 {object} common.Response "未授权"
// @Failure 500 {object} common.Response "内部服务器错误"
// @Router /api/oss/user/list [get]
func (c *UserController) ListUsers(ctx *gin.Context) {
	var req dto.UserListRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, common.ErrorResponse(err.Error()))
		return
	}

	users, err := c.userService.ListUsers(ctx, &req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, common.ErrorResponse(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, common.SuccessResponse(users))
}

// UpdateUserStatus 更新用户状态
// @Summary 更新用户状态
// @Description 更新指定用户的状态
// @Tags 用户模块
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer {{token}}"
// @Param id path int true "用户ID"
// @Param status query int true "状态：1-正常，2-禁用，3-锁定" Enums(1, 2, 3)
// @Success 200 {object} common.Response "成功"
// @Failure 400 {object} common.Response "请求参数错误"
// @Failure 401 {object} common.Response "未授权"
// @Failure 500 {object} common.Response "内部服务器错误"
// @Router /api/oss/user/status/{id} [get]
func (c *UserController) UpdateUserStatus(ctx *gin.Context) {
	// 解析用户ID
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, common.ErrorResponse("无效的用户ID"))
		return
	}

	// 解析状态
	statusStr := ctx.Query("status")
	status, err := strconv.Atoi(statusStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, common.ErrorResponse("无效的状态值"))
		return
	}

	err = c.userService.UpdateUserStatus(ctx, id, status)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, common.ErrorResponse(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, common.SuccessResponse(nil))
}

// GetUserRoles 获取用户角色
// @Summary 获取用户角色
// @Description 获取指定用户的所有角色
// @Tags 用户模块
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer {{token}}"
// @Param id path int true "用户ID"
// @Success 200 {object} common.Response{data=[]dto.RoleResponse} "成功"
// @Failure 400 {object} common.Response "请求参数错误"
// @Failure 401 {object} common.Response "未授权"
// @Failure 500 {object} common.Response "内部服务器错误"
// @Router /api/oss/user/roles/{id} [get]
func (c *UserController) GetUserRoles(ctx *gin.Context) {
	// 解析用户ID
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, common.ErrorResponse("无效的用户ID"))
		return
	}

	roles, err := c.userService.GetUserRoles(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, common.ErrorResponse(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, common.SuccessResponse(roles))
}

// AssignRoles 分配用户角色
// @Summary 分配用户角色
// @Description 为指定用户分配角色
// @Tags 用户模块
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer {{token}}"
// @Param id path int true "用户ID"
// @Param roleIds body []uint true "角色ID列表"
// @Success 200 {object} common.Response "成功"
// @Failure 400 {object} common.Response "请求参数错误"
// @Failure 401 {object} common.Response "未授权"
// @Failure 500 {object} common.Response "内部服务器错误"
// @Router /api/oss/user/roles/{id} [post]
func (c *UserController) AssignRoles(ctx *gin.Context) {
	// 解析用户ID
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, common.ErrorResponse("无效的用户ID"))
		return
	}

	// 解析角色ID列表
	var roleIDs []uint
	if err := ctx.ShouldBindJSON(&roleIDs); err != nil {
		ctx.JSON(http.StatusBadRequest, common.ErrorResponse("无效的角色ID列表"))
		return
	}

	err = c.userService.AssignRoles(ctx, id, roleIDs)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, common.ErrorResponse(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, common.SuccessResponse(nil))
}

// RemoveRoles 移除用户角色
// @Summary 移除用户角色
// @Description 移除指定用户的角色
// @Tags 用户模块
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer {{token}}"
// @Param id path int true "用户ID"
// @Param roleIds body []uint true "角色ID列表"
// @Success 200 {object} common.Response "成功"
// @Failure 400 {object} common.Response "请求参数错误"
// @Failure 401 {object} common.Response "未授权"
// @Failure 500 {object} common.Response "内部服务器错误"
// @Router /api/oss/user/roles/{id}/remove [post]
func (c *UserController) RemoveRoles(ctx *gin.Context) {
	// 解析用户ID
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, common.ErrorResponse("无效的用户ID"))
		return
	}

	// 解析角色ID列表
	var roleIDs []uint
	if err := ctx.ShouldBindJSON(&roleIDs); err != nil {
		ctx.JSON(http.StatusBadRequest, common.ErrorResponse("无效的角色ID列表"))
		return
	}

	err = c.userService.RemoveRoles(ctx, id, roleIDs)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, common.ErrorResponse(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, common.SuccessResponse(nil))
}
