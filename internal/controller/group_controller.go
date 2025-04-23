package controller

import (
	"net/http"
	"strconv"

	"oss-backend/internal/service"

	"github.com/gin-gonic/gin"
)

// GroupController 处理群组相关请求
type GroupController struct {
	groupService service.GroupService
}

// NewGroupController 创建新的群组控制器
func NewGroupController(groupService service.GroupService) *GroupController {
	return &GroupController{
		groupService: groupService,
	}
}

// CreateGroup 创建群组
// @Summary 创建群组
// @Description 创建一个新的群组
// @Tags group
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body service.CreateGroupRequest true "群组创建信息"
// @Success 201 {object} map[string]interface{} "创建成功"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Failure 409 {object} map[string]interface{} "群组已存在"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /group/create [post]
func (c *GroupController) CreateGroup(ctx *gin.Context) {
	var req service.CreateGroupRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误: " + err.Error(),
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

	// 创建群组
	group, err := c.groupService.CreateGroup(req.Name, req.Description, uint(userID.(uint)))
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err == service.ErrGroupNameExists {
			statusCode = http.StatusConflict
		}

		ctx.JSON(statusCode, gin.H{
			"code":    statusCode,
			"message": "创建群组失败: " + err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"code":    201,
		"message": "创建群组成功",
		"data":    group,
	})
}

// GetGroup 获取群组信息
// @Summary 获取群组信息
// @Description 获取指定群组的信息
// @Tags group
// @Produce json
// @Security BearerAuth
// @Param id path int true "群组ID"
// @Success 200 {object} map[string]interface{} "获取成功"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Failure 404 {object} map[string]interface{} "群组不存在"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /group/info/{id} [get]
func (c *GroupController) GetGroup(ctx *gin.Context) {
	groupID, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的群组ID",
		})
		return
	}

	// 获取群组信息
	group, err := c.groupService.GetGroupByID(uint(groupID))
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err == service.ErrGroupNotFound {
			statusCode = http.StatusNotFound
		}

		ctx.JSON(statusCode, gin.H{
			"code":    statusCode,
			"message": "获取群组信息失败: " + err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取群组信息成功",
		"data":    group,
	})
}

// UpdateGroup 更新群组信息
// @Summary 更新群组信息
// @Description 更新指定群组的信息
// @Tags group
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "群组ID"
// @Param request body service.UpdateGroupRequest true "群组更新信息"
// @Success 200 {object} map[string]interface{} "更新成功"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Failure 404 {object} map[string]interface{} "群组不存在"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /group/update/{id} [post]
func (c *GroupController) UpdateGroup(ctx *gin.Context) {
	groupID, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的群组ID",
		})
		return
	}

	var req service.UpdateGroupRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误: " + err.Error(),
		})
		return
	}

	// 更新群组信息
	group, err := c.groupService.UpdateGroup(uint(groupID), req.Name, req.Description)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err == service.ErrGroupNotFound {
			statusCode = http.StatusNotFound
		} else if err == service.ErrGroupNameExists {
			statusCode = http.StatusConflict
		}

		ctx.JSON(statusCode, gin.H{
			"code":    statusCode,
			"message": "更新群组信息失败: " + err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "更新群组信息成功",
		"data":    group,
	})
}

// DeleteGroup 删除群组
// @Summary 删除群组
// @Description 删除指定的群组
// @Tags group
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "群组ID"
// @Success 200 {object} map[string]interface{} "删除成功"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Failure 404 {object} map[string]interface{} "群组不存在"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /group/delete/{id} [post]
func (c *GroupController) DeleteGroup(ctx *gin.Context) {
	groupID, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的群组ID",
		})
		return
	}

	// 删除群组
	err = c.groupService.DeleteGroup(uint(groupID))
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err == service.ErrGroupNotFound {
			statusCode = http.StatusNotFound
		} else if err == service.ErrGroupNotEmpty {
			statusCode = http.StatusBadRequest
		}

		ctx.JSON(statusCode, gin.H{
			"code":    statusCode,
			"message": "删除群组失败: " + err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "删除群组成功",
	})
}

// ListGroups 获取群组列表
// @Summary 获取用户群组列表
// @Description 获取当前用户所在的所有群组
// @Tags group
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "获取成功"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /group/list [get]
func (c *GroupController) ListGroups(ctx *gin.Context) {
	// 获取当前用户ID
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "未授权：用户未登录",
		})
		return
	}

	// 获取用户的群组列表
	groups, err := c.groupService.GetUserGroups(uint(userID.(uint)))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取群组列表失败: " + err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取群组列表成功",
		"data":    groups,
	})
}

// GetGroupMembers 获取群组成员列表
// @Summary 获取群组成员列表
// @Description 获取指定群组的所有成员
// @Tags group
// @Produce json
// @Security BearerAuth
// @Param id path int true "群组ID"
// @Success 200 {object} map[string]interface{} "获取成功"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Failure 404 {object} map[string]interface{} "群组不存在"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /group/members/{id} [get]
func (c *GroupController) GetGroupMembers(ctx *gin.Context) {
	groupID, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的群组ID",
		})
		return
	}

	// 获取群组成员列表
	members, err := c.groupService.GetGroupMembers(uint(groupID))
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err == service.ErrGroupNotFound {
			statusCode = http.StatusNotFound
		}

		ctx.JSON(statusCode, gin.H{
			"code":    statusCode,
			"message": "获取群组成员列表失败: " + err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取群组成员列表成功",
		"data":    members,
	})
}

// AddGroupMember 添加群组成员
// @Summary 添加群组成员
// @Description 添加成员到指定群组
// @Tags group
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "群组ID"
// @Param request body service.AddMemberRequest true "添加成员请求"
// @Success 200 {object} map[string]interface{} "添加成功"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Failure 404 {object} map[string]interface{} "群组或用户不存在"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /group/member/add/{id} [post]
func (c *GroupController) AddGroupMember(ctx *gin.Context) {
	groupID, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的群组ID",
		})
		return
	}

	var req service.AddMemberRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误: " + err.Error(),
		})
		return
	}

	// 添加群组成员
	err = c.groupService.AddGroupMember(uint(groupID), uint(req.UserID), req.Role)
	if err != nil {
		statusCode := http.StatusInternalServerError

		ctx.JSON(statusCode, gin.H{
			"code":    statusCode,
			"message": "添加群组成员失败: " + err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "添加群组成员成功",
	})
}

// RemoveGroupMember 移除群组成员
// @Summary 移除群组成员
// @Description 从指定群组中移除成员
// @Tags group
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "群组ID"
// @Param user_id path int true "用户ID"
// @Success 200 {object} map[string]interface{} "移除成功"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Failure 404 {object} map[string]interface{} "群组或用户不存在"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /group/member/remove/{id}/{user_id} [post]
func (c *GroupController) RemoveGroupMember(ctx *gin.Context) {
	groupID, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的群组ID",
		})
		return
	}

	userID, err := strconv.ParseUint(ctx.Param("user_id"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的用户ID",
		})
		return
	}

	// 移除群组成员
	err = c.groupService.RemoveGroupMember(uint(groupID), uint(userID))
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err == service.ErrGroupNotFound {
			statusCode = http.StatusNotFound
		} else if err == service.ErrCannotRemoveOwner {
			statusCode = http.StatusBadRequest
		}

		ctx.JSON(statusCode, gin.H{
			"code":    statusCode,
			"message": "移除群组成员失败: " + err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "移除群组成员成功",
	})
}

// UpdateMemberRole 更新成员角色
// @Summary 更新成员角色
// @Description 更新指定群组成员的角色
// @Tags group
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "群组ID"
// @Param user_id path int true "用户ID"
// @Param request body struct { Role string } true "更新角色请求"
// @Success 200 {object} map[string]interface{} "更新成功"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Failure 404 {object} map[string]interface{} "群组或用户不存在"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /group/member/update-role/{id}/{user_id} [post]
func (c *GroupController) UpdateMemberRole(ctx *gin.Context) {
	groupID, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的群组ID",
		})
		return
	}

	userID, err := strconv.ParseUint(ctx.Param("user_id"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的用户ID",
		})
		return
	}

	var req struct {
		Role string `json:"role" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误: " + err.Error(),
		})
		return
	}

	// 更新成员角色
	err = c.groupService.UpdateMemberRole(uint(groupID), uint(userID), req.Role)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err == service.ErrGroupNotFound || err == service.ErrNotGroupMember {
			statusCode = http.StatusNotFound
		} else if err == service.ErrInsufficientPermission {
			statusCode = http.StatusForbidden
		}

		ctx.JSON(statusCode, gin.H{
			"code":    statusCode,
			"message": "更新成员角色失败: " + err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "更新成员角色成功",
	})
}
