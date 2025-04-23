package controller

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"oss-backend/internal/model/dto"
	"oss-backend/internal/service"
	"oss-backend/pkg/common"
)

// RoleController 角色控制器
type RoleController struct {
	roleService service.RoleService
}

// NewRoleController 创建角色控制器
func NewRoleController(roleService service.RoleService) *RoleController {
	return &RoleController{
		roleService: roleService,
	}
}

// CreateRole 创建角色
// @Summary 创建角色
// @Description 创建一个新角色
// @Tags 角色管理
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer 用户令牌"
// @Param request body dto.RoleCreateRequest true "角色信息"
// @Success 200 {object} common.Response "成功"
// @Failure 400 {object} common.Response "请求参数错误"
// @Failure 401 {object} common.Response "未授权"
// @Failure 500 {object} common.Response "内部服务器错误"
// @Router /api/oss/role/create [post]
func (c *RoleController) CreateRole(ctx *gin.Context) {
	var req dto.RoleCreateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, common.ErrorResponse(err.Error()))
		return
	}

	// 获取当前用户ID，实际项目中应该从JWT Token中获取
	var createdBy uint = 1 // 临时写死，实际项目中需要从Token中获取

	err := c.roleService.CreateRole(ctx, &req, createdBy)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, common.ErrorResponse(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, common.SuccessResponse(nil))
}

// UpdateRole 更新角色
// @Summary 更新角色
// @Description 更新角色信息
// @Tags 角色管理
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer 用户令牌"
// @Param request body dto.RoleUpdateRequest true "角色信息"
// @Success 200 {object} common.Response "成功"
// @Failure 400 {object} common.Response "请求参数错误"
// @Failure 401 {object} common.Response "未授权"
// @Failure 500 {object} common.Response "内部服务器错误"
// @Router /api/oss/role/update [post]
func (c *RoleController) UpdateRole(ctx *gin.Context) {
	var req dto.RoleUpdateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, common.ErrorResponse(err.Error()))
		return
	}

	// 获取当前用户ID，实际项目中应该从JWT Token中获取
	var updatedBy uint = 1 // 临时写死，实际项目中需要从Token中获取

	err := c.roleService.UpdateRole(ctx, &req, updatedBy)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, common.ErrorResponse(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, common.SuccessResponse(nil))
}

// DeleteRole 删除角色
// @Summary 删除角色
// @Description 删除指定ID的角色
// @Tags 角色管理
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer 用户令牌"
// @Param id path int true "角色ID"
// @Success 200 {object} common.Response "成功"
// @Failure 400 {object} common.Response "请求参数错误"
// @Failure 401 {object} common.Response "未授权"
// @Failure 500 {object} common.Response "内部服务器错误"
// @Router /api/oss/role/delete/{id} [get]
func (c *RoleController) DeleteRole(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, common.ErrorResponse("无效的角色ID"))
		return
	}

	err = c.roleService.DeleteRole(ctx, uint(id))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, common.ErrorResponse(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, common.SuccessResponse(nil))
}

// GetRoleByID 根据ID获取角色
// @Summary 获取角色详情
// @Description 根据ID获取角色详情
// @Tags 角色管理
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer 用户令牌"
// @Param id path int true "角色ID"
// @Success 200 {object} common.Response{data=dto.RoleResponse} "成功"
// @Failure 400 {object} common.Response "请求参数错误"
// @Failure 401 {object} common.Response "未授权"
// @Failure 500 {object} common.Response "内部服务器错误"
// @Router /api/oss/role/detail/{id} [get]
func (c *RoleController) GetRoleByID(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, common.ErrorResponse("无效的角色ID"))
		return
	}

	role, err := c.roleService.GetRoleByID(ctx, uint(id))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, common.ErrorResponse(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, common.SuccessResponse(role))
}

// ListRoles 获取角色列表
// @Summary 获取角色列表
// @Description 根据条件获取角色列表
// @Tags 角色管理
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer 用户令牌"
// @Param name query string false "角色名称，模糊查询"
// @Param status query int false "状态：1-启用，0-禁用" Enums(0, 1)
// @Param page query int false "页码，默认1"
// @Param size query int false "每页数量，默认10"
// @Success 200 {object} common.Response{data=dto.RoleListResponse} "成功"
// @Failure 400 {object} common.Response "请求参数错误"
// @Failure 401 {object} common.Response "未授权"
// @Failure 500 {object} common.Response "内部服务器错误"
// @Router /api/oss/role/list [get]
func (c *RoleController) ListRoles(ctx *gin.Context) {
	var req dto.RoleListRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, common.ErrorResponse(err.Error()))
		return
	}

	roles, err := c.roleService.ListRoles(ctx, &req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, common.ErrorResponse(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, common.SuccessResponse(roles))
}
