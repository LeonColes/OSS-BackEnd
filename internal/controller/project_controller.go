package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"oss-backend/internal/model/dto"
	"oss-backend/internal/service"
	"oss-backend/pkg/common"
)

// ProjectController 项目控制器
type ProjectController struct {
	projectService service.ProjectService
}

// NewProjectController 创建项目控制器
func NewProjectController(projectService service.ProjectService) *ProjectController {
	return &ProjectController{
		projectService: projectService,
	}
}

// CreateProject 创建项目
// @Summary 创建项目
// @Description 创建一个新项目（需要组管理员或系统管理员权限）
// @Tags 项目管理
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer {{token}}"
// @Param request body dto.CreateProjectRequest true "项目信息"
// @Success 200 {object} common.Response{data=dto.ProjectResponse} "成功"
// @Failure 400 {object} common.Response "请求参数错误"
// @Failure 401 {object} common.Response "未授权"
// @Failure 403 {object} common.Response "无权限"
// @Failure 500 {object} common.Response "服务器内部错误"
// @Router /api/oss/project/create [post]
func (c *ProjectController) CreateProject(ctx *gin.Context) {
	// 获取当前用户ID
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, common.ErrorResponse("未授权"))
		return
	}

	// 解析请求参数
	var req dto.CreateProjectRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, common.ErrorResponse("请求参数错误: "+err.Error()))
		return
	}

	// 调用服务创建项目
	project, err := c.projectService.CreateProject(ctx, &req, userID.(string))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, common.ErrorResponse("创建项目失败: "+err.Error()))
		return
	}

	// 返回成功响应
	ctx.JSON(http.StatusOK, common.SuccessResponse(project))
}

// UpdateProject 更新项目
// @Summary 更新项目
// @Description 更新项目信息（需要项目管理员权限）
// @Tags 项目管理
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer {{token}}"
// @Param request body dto.UpdateProjectRequest true "项目信息"
// @Success 200 {object} common.Response{data=dto.ProjectResponse} "成功"
// @Failure 400 {object} common.Response "请求参数错误"
// @Failure 401 {object} common.Response "未授权"
// @Failure 403 {object} common.Response "无权限"
// @Failure 500 {object} common.Response "服务器内部错误"
// @Router /api/oss/project/update [post]
func (c *ProjectController) UpdateProject(ctx *gin.Context) {
	// 获取当前用户ID
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, common.ErrorResponse("未授权"))
		return
	}

	// 解析请求参数
	var req dto.UpdateProjectRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, common.ErrorResponse("请求参数错误: "+err.Error()))
		return
	}

	// 调用服务更新项目
	project, err := c.projectService.UpdateProject(ctx, &req, userID.(string))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, common.ErrorResponse("更新项目失败: "+err.Error()))
		return
	}

	// 返回成功响应
	ctx.JSON(http.StatusOK, common.SuccessResponse(project))
}

// GetProjectByID 获取项目详情
// @Summary 获取项目详情
// @Description 获取项目详细信息
// @Tags 项目管理
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer {{token}}"
// @Param id path int true "项目ID"
// @Success 200 {object} common.Response{data=dto.ProjectResponse} "成功"
// @Failure 400 {object} common.Response "请求参数错误"
// @Failure 401 {object} common.Response "未授权"
// @Failure 403 {object} common.Response "无权限"
// @Failure 500 {object} common.Response "服务器内部错误"
// @Router /api/oss/project/detail/{id} [get]
func (c *ProjectController) GetProjectByID(ctx *gin.Context) {
	// 获取当前用户ID
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, common.ErrorResponse("未授权"))
		return
	}

	// 获取项目ID
	projectIDStr := ctx.Param("id")
	projectID := projectIDStr

	// 调用服务获取项目详情
	project, err := c.projectService.GetProjectByID(ctx, projectID, userID.(string))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, common.ErrorResponse("获取项目详情失败: "+err.Error()))
		return
	}

	// 返回成功响应
	ctx.JSON(http.StatusOK, common.SuccessResponse(project))
}

// ListProjects 列出项目
// @Summary 列出项目
// @Description 获取项目列表
// @Tags 项目管理
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer {{token}}"
// @Param group_id query int false "群组ID"
// @Param status query int false "项目状态: 1-正常, 2-归档, 3-删除"
// @Param keyword query string false "关键词"
// @Param page query int false "页码"
// @Param size query int false "每页大小"
// @Success 200 {object} common.Response{data=common.PageResult{list=[]dto.ProjectResponse}} "成功"
// @Failure 401 {object} common.Response "未授权"
// @Failure 500 {object} common.Response "服务器内部错误"
// @Router /api/oss/project/list [get]
func (c *ProjectController) ListProjects(ctx *gin.Context) {
	// 获取当前用户ID
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, common.ErrorResponse("未授权"))
		return
	}

	// 解析查询参数
	var query dto.ProjectQuery
	if err := ctx.ShouldBindQuery(&query); err != nil {
		ctx.JSON(http.StatusBadRequest, common.ErrorResponse("请求参数错误: "+err.Error()))
		return
	}

	// 调用服务获取项目列表
	projects, total, err := c.projectService.ListProjects(ctx, &query, userID.(string))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, common.ErrorResponse("获取项目列表失败: "+err.Error()))
		return
	}

	// 返回成功响应
	ctx.JSON(http.StatusOK, common.SuccessResponse(common.PageResult{
		Total: total,
		List:  projects,
	}))
}

// GetUserProjects 获取用户参与的项目
// @Summary 获取用户参与的项目
// @Description 获取当前用户参与的项目列表
// @Tags 项目管理
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer {{token}}"
// @Param group_id query int false "群组ID"
// @Param status query int false "项目状态: 1-正常, 2-归档, 3-删除"
// @Param keyword query string false "关键词"
// @Param page query int false "页码"
// @Param size query int false "每页大小"
// @Success 200 {object} common.Response{data=common.PageResult{list=[]dto.ProjectResponse}} "成功"
// @Failure 401 {object} common.Response "未授权"
// @Failure 500 {object} common.Response "服务器内部错误"
// @Router /api/oss/project/user [get]
func (c *ProjectController) GetUserProjects(ctx *gin.Context) {
	// 获取当前用户ID
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, common.ErrorResponse("未授权"))
		return
	}

	// 解析查询参数
	var query dto.ProjectQuery
	if err := ctx.ShouldBindQuery(&query); err != nil {
		ctx.JSON(http.StatusBadRequest, common.ErrorResponse("请求参数错误: "+err.Error()))
		return
	}

	// 调用服务获取用户参与的项目
	projects, total, err := c.projectService.GetUserProjects(ctx, &query, userID.(string))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, common.ErrorResponse("获取用户项目失败: "+err.Error()))
		return
	}

	// 返回成功响应
	ctx.JSON(http.StatusOK, common.SuccessResponse(common.PageResult{
		Total: total,
		List:  projects,
	}))
}

// DeleteProject 删除项目
// @Summary 删除项目
// @Description 逻辑删除项目
// @Tags 项目管理
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer {{token}}"
// @Param id path int true "项目ID"
// @Success 200 {object} common.Response "成功"
// @Failure 400 {object} common.Response "请求参数错误"
// @Failure 401 {object} common.Response "未授权"
// @Failure 403 {object} common.Response "无权限"
// @Failure 500 {object} common.Response "服务器内部错误"
// @Router /api/oss/project/delete/{id} [get]
func (c *ProjectController) DeleteProject(ctx *gin.Context) {
	// 获取当前用户ID
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, common.ErrorResponse("未授权"))
		return
	}

	// 获取项目ID
	projectIDStr := ctx.Param("id")
	projectID := projectIDStr

	// 调用服务删除项目
	err := c.projectService.DeleteProject(ctx, projectID, userID.(string))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, common.ErrorResponse("删除项目失败: "+err.Error()))
		return
	}

	// 返回成功响应
	ctx.JSON(http.StatusOK, common.SuccessResponse(nil))
}

// SetPermission 设置项目成员权限
// @Summary 设置项目成员权限
// @Description 为项目成员设置权限（需要项目管理员权限）
// @Tags 项目管理
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer {{token}}"
// @Param request body dto.SetPermissionRequest true "权限信息"
// @Success 200 {object} common.Response "成功"
// @Failure 400 {object} common.Response "请求参数错误"
// @Failure 401 {object} common.Response "未授权"
// @Failure 403 {object} common.Response "无权限"
// @Failure 500 {object} common.Response "服务器内部错误"
// @Router /api/oss/project/permission/set [post]
func (c *ProjectController) SetPermission(ctx *gin.Context) {
	// 获取当前用户ID
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, common.ErrorResponse("未授权"))
		return
	}

	// 解析请求参数
	var req dto.SetPermissionRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, common.ErrorResponse("请求参数错误: "+err.Error()))
		return
	}

	// 调用服务设置权限
	err := c.projectService.SetPermission(ctx, &req, userID.(string))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, common.ErrorResponse("设置项目权限失败: "+err.Error()))
		return
	}

	// 返回成功响应
	ctx.JSON(http.StatusOK, common.SuccessResponse(nil))
}

// RemovePermission 移除项目成员权限
// @Summary 移除项目成员权限
// @Description 移除项目成员的权限（需要项目管理员权限）
// @Tags 项目管理
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer {{token}}"
// @Param request body dto.RemovePermissionRequest true "移除信息"
// @Success 200 {object} common.Response "成功"
// @Failure 400 {object} common.Response "请求参数错误"
// @Failure 401 {object} common.Response "未授权"
// @Failure 403 {object} common.Response "无权限"
// @Failure 500 {object} common.Response "服务器内部错误"
// @Router /api/oss/project/permission/remove [post]
func (c *ProjectController) RemovePermission(ctx *gin.Context) {
	// 获取当前用户ID
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, common.ErrorResponse("未授权"))
		return
	}

	// 解析请求参数
	var req dto.RemovePermissionRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, common.ErrorResponse("请求参数错误: "+err.Error()))
		return
	}

	// 调用服务移除权限
	err := c.projectService.RemovePermission(ctx, &req, userID.(string))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, common.ErrorResponse("移除项目权限失败: "+err.Error()))
		return
	}

	// 返回成功响应
	ctx.JSON(http.StatusOK, common.SuccessResponse(nil))
}

// ListProjectUsers 列出项目成员
// @Summary 列出项目成员
// @Description 获取项目所有成员及其权限
// @Tags 项目管理
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer {{token}}"
// @Param id path int true "项目ID"
// @Success 200 {object} common.Response{data=[]dto.ProjectUserResponse} "成功"
// @Failure 400 {object} common.Response "请求参数错误"
// @Failure 401 {object} common.Response "未授权"
// @Failure 403 {object} common.Response "无权限"
// @Failure 500 {object} common.Response "服务器内部错误"
// @Router /api/oss/project/users/{id} [get]
func (c *ProjectController) ListProjectUsers(ctx *gin.Context) {
	// 获取当前用户ID
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, common.ErrorResponse("未授权"))
		return
	}

	// 获取项目ID
	projectIDStr := ctx.Param("id")
	projectID := projectIDStr

	// 调用服务获取项目成员
	users, err := c.projectService.ListProjectUsers(ctx, projectID, userID.(string))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, common.ErrorResponse("获取项目成员失败: "+err.Error()))
		return
	}

	// 返回成功响应
	ctx.JSON(http.StatusOK, common.SuccessResponse(users))
}
