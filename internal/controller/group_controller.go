package controller

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"oss-backend/internal/model/dto"
	"oss-backend/internal/service"
	"oss-backend/pkg/common"
)

// GroupController 群组控制器
type GroupController struct {
	groupService service.GroupService
}

// NewGroupController 创建群组控制器
func NewGroupController(groupService service.GroupService) *GroupController {
	return &GroupController{
		groupService: groupService,
	}
}

// CreateGroup 创建群组
// @Summary 创建群组
// @Description 创建一个新的群组
// @Tags 群组管理
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer {{token}}"
// @Param request body dto.GroupCreateRequest true "群组信息"
// @Success 200 {object} common.Response "成功"
// @Failure 400 {object} common.Response "请求参数错误"
// @Failure 401 {object} common.Response "未授权"
// @Failure 500 {object} common.Response "内部服务器错误"
// @Router /api/oss/group/create [post]
func (c *GroupController) CreateGroup(ctx *gin.Context) {
	var req dto.GroupCreateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, common.ErrorResponse(err.Error()))
		return
	}

	// 获取当前用户ID
	userIDValue, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, common.ErrorResponse("未授权"))
		return
	}
	userID := userIDValue.(string)

	err := c.groupService.CreateGroup(ctx, &req, userID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, common.ErrorResponse(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, common.SuccessResponse(nil))
}

// UpdateGroup 更新群组
// @Summary 更新群组
// @Description 更新群组信息
// @Tags 群组管理
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer {{token}}"
// @Param request body dto.GroupUpdateRequest true "群组信息"
// @Success 200 {object} common.Response "成功"
// @Failure 400 {object} common.Response "请求参数错误"
// @Failure 401 {object} common.Response "未授权"
// @Failure 500 {object} common.Response "内部服务器错误"
// @Router /api/oss/group/update [post]
func (c *GroupController) UpdateGroup(ctx *gin.Context) {
	var req dto.GroupUpdateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, common.ErrorResponse(err.Error()))
		return
	}

	// 获取当前用户ID
	userIDValue, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, common.ErrorResponse("未授权"))
		return
	}
	userID := userIDValue.(string)

	err := c.groupService.UpdateGroup(ctx, &req, userID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, common.ErrorResponse(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, common.SuccessResponse(nil))
}

// GetGroupByID 获取群组详情
// @Summary 获取群组详情
// @Description 根据ID获取群组详情
// @Tags 群组管理
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer {{token}}"
// @Param id path int true "群组ID"
// @Success 200 {object} common.Response{data=dto.GroupResponse} "成功"
// @Failure 400 {object} common.Response "请求参数错误"
// @Failure 401 {object} common.Response "未授权"
// @Failure 500 {object} common.Response "内部服务器错误"
// @Router /api/oss/group/detail/{id} [get]
func (c *GroupController) GetGroupByID(ctx *gin.Context) {
	// 解析群组ID
	idStr := ctx.Param("id")
	id := idStr

	// 获取当前用户ID
	userIDValue, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, common.ErrorResponse("未授权"))
		return
	}
	userID := userIDValue.(string)

	group, err := c.groupService.GetGroupByID(ctx, id, userID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, common.ErrorResponse(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, common.SuccessResponse(group))
}

// ListGroups 获取群组列表
// @Summary 获取群组列表
// @Description 根据条件获取群组列表
// @Tags 群组管理
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer {{token}}"
// @Param name query string false "群组名称，模糊查询"
// @Param status query int false "状态：1-正常，2-禁用，3-锁定"
// @Param page query int false "页码，默认1"
// @Param size query int false "每页数量，默认10"
// @Success 200 {object} common.Response{data=dto.GroupListResponse} "成功"
// @Failure 400 {object} common.Response "请求参数错误"
// @Failure 401 {object} common.Response "未授权"
// @Failure 500 {object} common.Response "内部服务器错误"
// @Router /api/oss/group/list [get]
func (c *GroupController) ListGroups(ctx *gin.Context) {
	var req dto.GroupListRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, common.ErrorResponse(err.Error()))
		return
	}

	// 获取当前用户ID
	userIDValue, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, common.ErrorResponse("未授权"))
		return
	}
	userID := userIDValue.(string)

	groups, err := c.groupService.ListGroups(ctx, &req, userID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, common.ErrorResponse(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, common.SuccessResponse(groups))
}

// JoinGroup 加入群组
// @Summary 加入群组
// @Description 通过邀请码加入群组
// @Tags 群组管理
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer {{token}}"
// @Param request body dto.GroupJoinRequest true "加入信息"
// @Success 200 {object} common.Response "成功"
// @Failure 400 {object} common.Response "请求参数错误"
// @Failure 401 {object} common.Response "未授权"
// @Failure 500 {object} common.Response "内部服务器错误"
// @Router /api/oss/group/join [post]
func (c *GroupController) JoinGroup(ctx *gin.Context) {
	var req dto.GroupJoinRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, common.ErrorResponse(err.Error()))
		return
	}

	// 获取当前用户ID
	userIDValue, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, common.ErrorResponse("未授权"))
		return
	}
	userID := userIDValue.(string)

	err := c.groupService.JoinGroup(ctx, &req, userID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, common.ErrorResponse(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, common.SuccessResponse(nil))
}

// AddMember 添加成员
// @Summary 添加成员
// @Description 向群组添加成员
// @Tags 群组管理
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer {{token}}"
// @Param id path int true "群组ID"
// @Param user_id query int true "用户ID"
// @Param role query string true "角色：admin, member" Enums(admin, member)
// @Success 200 {object} common.Response "成功"
// @Failure 400 {object} common.Response "请求参数错误"
// @Failure 401 {object} common.Response "未授权"
// @Failure 500 {object} common.Response "内部服务器错误"
// @Router /api/oss/group/member/add/{id} [get]
func (c *GroupController) AddMember(ctx *gin.Context) {
	// 解析群组ID
	idStr := ctx.Param("id")
	groupID := idStr

	// 解析用户ID
	userIDStr := ctx.Query("user_id")
	userID := userIDStr

	// 解析角色
	role := ctx.Query("role")
	if role != "admin" && role != "member" {
		ctx.JSON(http.StatusBadRequest, common.ErrorResponse("无效的角色"))
		return
	}

	// 获取当前用户ID
	operatorIDValue, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, common.ErrorResponse("未授权"))
		return
	}
	operatorID := operatorIDValue.(string)

	err := c.groupService.AddMember(ctx, groupID, userID, role, operatorID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, common.ErrorResponse(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, common.SuccessResponse(nil))
}

// UpdateMemberRole 更新成员角色
// @Summary 更新成员角色
// @Description 更新群组成员的角色
// @Tags 群组管理
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer {{token}}"
// @Param id path int true "群组ID"
// @Param request body dto.GroupMemberUpdateRequest true "成员信息"
// @Success 200 {object} common.Response "成功"
// @Failure 400 {object} common.Response "请求参数错误"
// @Failure 401 {object} common.Response "未授权"
// @Failure 500 {object} common.Response "内部服务器错误"
// @Router /api/oss/group/member/role/{id} [post]
func (c *GroupController) UpdateMemberRole(ctx *gin.Context) {
	// 解析群组ID
	idStr := ctx.Param("id")
	groupID := idStr

	var req dto.GroupMemberUpdateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, common.ErrorResponse(err.Error()))
		return
	}

	// 获取当前用户ID
	operatorIDValue, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, common.ErrorResponse("未授权"))
		return
	}
	operatorID := operatorIDValue.(string)

	err := c.groupService.UpdateMemberRole(ctx, groupID, &req, operatorID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, common.ErrorResponse(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, common.SuccessResponse(nil))
}

// RemoveMember 移除成员
// @Summary 移除成员
// @Description 从群组中移除成员
// @Tags 群组管理
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer {{token}}"
// @Param id path int true "群组ID"
// @Param user_id query int true "用户ID"
// @Success 200 {object} common.Response "成功"
// @Failure 400 {object} common.Response "请求参数错误"
// @Failure 401 {object} common.Response "未授权"
// @Failure 500 {object} common.Response "内部服务器错误"
// @Router /api/oss/group/member/remove/{id} [get]
func (c *GroupController) RemoveMember(ctx *gin.Context) {
	// 解析群组ID
	idStr := ctx.Param("id")
	groupID := idStr

	// 解析用户ID
	userIDStr := ctx.Query("user_id")
	userID := userIDStr

	// 获取当前用户ID
	operatorIDValue, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, common.ErrorResponse("未授权"))
		return
	}
	operatorID := operatorIDValue.(string)

	err := c.groupService.RemoveMember(ctx, groupID, userID, operatorID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, common.ErrorResponse(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, common.SuccessResponse(nil))
}

// ListMembers 获取成员列表
// @Summary 获取成员列表
// @Description 获取群组成员列表
// @Tags 群组管理
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer {{token}}"
// @Param id path int true "群组ID"
// @Param page query int false "页码，默认1"
// @Param size query int false "每页数量，默认10"
// @Success 200 {object} common.Response{data=dto.GroupMemberListResponse} "成功"
// @Failure 400 {object} common.Response "请求参数错误"
// @Failure 401 {object} common.Response "未授权"
// @Failure 500 {object} common.Response "内部服务器错误"
// @Router /api/oss/group/member/list/{id} [get]
func (c *GroupController) ListMembers(ctx *gin.Context) {
	// 解析群组ID
	idStr := ctx.Param("id")
	groupID := idStr

	// 解析分页参数
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(ctx.DefaultQuery("size", "10"))

	members, err := c.groupService.ListMembers(ctx, groupID, page, size)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, common.ErrorResponse(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, common.SuccessResponse(members))
}

// GetUserGroups 获取用户所属的群组
// @Summary 获取用户所属的群组
// @Description 获取当前用户所属的所有群组
// @Tags 群组管理
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer {{token}}"
// @Success 200 {object} common.Response{data=[]dto.GroupResponse} "成功"
// @Failure 401 {object} common.Response "未授权"
// @Failure 500 {object} common.Response "内部服务器错误"
// @Router /api/oss/group/user [get]
func (c *GroupController) GetUserGroups(ctx *gin.Context) {
	// 获取当前用户ID
	userIDValue, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, common.ErrorResponse("未授权"))
		return
	}
	userID := userIDValue.(string)

	groups, err := c.groupService.GetUserGroups(ctx, userID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, common.ErrorResponse(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, common.SuccessResponse(groups))
}

// GenerateInviteCode 生成邀请码
// @Summary 生成邀请码
// @Description 为群组生成新的邀请码
// @Tags 群组管理
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer {{token}}"
// @Param request body dto.GroupInviteRequest true "邀请码信息"
// @Success 200 {object} common.Response{data=dto.GroupInviteResponse} "成功"
// @Failure 400 {object} common.Response "请求参数错误"
// @Failure 401 {object} common.Response "未授权"
// @Failure 500 {object} common.Response "内部服务器错误"
// @Router /api/oss/group/invite [post]
func (c *GroupController) GenerateInviteCode(ctx *gin.Context) {
	var req dto.GroupInviteRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, common.ErrorResponse(err.Error()))
		return
	}

	// 获取当前用户ID
	userIDValue, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, common.ErrorResponse("未授权"))
		return
	}
	userID := userIDValue.(string)

	invite, err := c.groupService.GenerateInviteCode(ctx, &req, userID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, common.ErrorResponse(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, common.SuccessResponse(invite))
}
