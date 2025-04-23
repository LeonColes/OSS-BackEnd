package controller

import (
	"net/http"
	"strconv"

	"oss-backend/internal/service"

	"github.com/gin-gonic/gin"
)

// ProjectController 处理项目相关请求
type ProjectController struct {
	projectService service.ProjectService
}

// NewProjectController 创建新的项目控制器
func NewProjectController(projectService service.ProjectService) *ProjectController {
	return &ProjectController{
		projectService: projectService,
	}
}

// CreateProject 创建项目
func (c *ProjectController) CreateProject(ctx *gin.Context) {
	var req struct {
		GroupID     uint   `json:"group_id" binding:"required"`
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
		PathPrefix  string `json:"path_prefix"`
	}

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

	// 创建项目
	project, err := c.projectService.CreateProject(
		req.GroupID,
		req.Name,
		req.Description,
		req.PathPrefix,
		uint(userID.(uint)),
	)

	if err != nil {
		statusCode := http.StatusInternalServerError
		if err == service.ErrNoProjectPermission {
			statusCode = http.StatusForbidden
		} else if err == service.ErrPathPrefixExists {
			statusCode = http.StatusConflict
		}

		ctx.JSON(statusCode, gin.H{
			"code":    statusCode,
			"message": "创建项目失败: " + err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"code":    201,
		"message": "创建项目成功",
		"data":    project,
	})
}

// GetProject 获取项目信息
func (c *ProjectController) GetProject(ctx *gin.Context) {
	projectID, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的项目ID",
		})
		return
	}

	// 获取项目信息
	project, err := c.projectService.GetProject(uint(projectID))
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err == service.ErrProjectNotFound {
			statusCode = http.StatusNotFound
		}

		ctx.JSON(statusCode, gin.H{
			"code":    statusCode,
			"message": "获取项目信息失败: " + err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取项目信息成功",
		"data":    project,
	})
}

// UpdateProject 更新项目信息
func (c *ProjectController) UpdateProject(ctx *gin.Context) {
	projectID, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的项目ID",
		})
		return
	}

	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}

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

	// 更新项目信息
	project, err := c.projectService.UpdateProject(
		uint(projectID),
		req.Name,
		req.Description,
		uint(userID.(uint)),
	)

	if err != nil {
		statusCode := http.StatusInternalServerError
		if err == service.ErrProjectNotFound {
			statusCode = http.StatusNotFound
		} else if err == service.ErrNoProjectPermission {
			statusCode = http.StatusForbidden
		}

		ctx.JSON(statusCode, gin.H{
			"code":    statusCode,
			"message": "更新项目信息失败: " + err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "更新项目信息成功",
		"data":    project,
	})
}

// DeleteProject 删除项目
func (c *ProjectController) DeleteProject(ctx *gin.Context) {
	projectID, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的项目ID",
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

	// 删除项目
	err = c.projectService.DeleteProject(uint(projectID), uint(userID.(uint)))
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err == service.ErrProjectNotFound {
			statusCode = http.StatusNotFound
		} else if err == service.ErrNoProjectPermission {
			statusCode = http.StatusForbidden
		}

		ctx.JSON(statusCode, gin.H{
			"code":    statusCode,
			"message": "删除项目失败: " + err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "删除项目成功",
	})
}

// ListProjects 列出项目
func (c *ProjectController) ListProjects(ctx *gin.Context) {
	// 获取当前用户ID
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "未授权：用户未登录",
		})
		return
	}

	// 获取用户的项目列表
	projects, err := c.projectService.GetUserProjects(uint(userID.(uint)))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取项目列表失败: " + err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取项目列表成功",
		"data":    projects,
	})
}

// GetProjectMembers 获取项目成员列表
func (c *ProjectController) GetProjectMembers(ctx *gin.Context) {
	projectID, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的项目ID",
		})
		return
	}

	// 获取项目成员列表
	members, err := c.projectService.GetProjectMembers(uint(projectID))
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err == service.ErrProjectNotFound {
			statusCode = http.StatusNotFound
		}

		ctx.JSON(statusCode, gin.H{
			"code":    statusCode,
			"message": "获取项目成员列表失败: " + err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取项目成员列表成功",
		"data":    members,
	})
}

// AddProjectMember 添加项目成员
func (c *ProjectController) AddProjectMember(ctx *gin.Context) {
	projectID, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的项目ID",
		})
		return
	}

	var req struct {
		UserID uint   `json:"user_id" binding:"required"`
		Role   string `json:"role" binding:"required"`
	}

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

	// 添加项目成员
	err = c.projectService.AddProjectMember(
		uint(projectID),
		req.UserID,
		req.Role,
		uint(userID.(uint)),
	)

	if err != nil {
		statusCode := http.StatusInternalServerError
		if err == service.ErrProjectNotFound || err == service.ErrUserNotFound {
			statusCode = http.StatusNotFound
		} else if err == service.ErrNoProjectPermission {
			statusCode = http.StatusForbidden
		} else if err == service.ErrMemberAlreadyExists {
			statusCode = http.StatusConflict
		}

		ctx.JSON(statusCode, gin.H{
			"code":    statusCode,
			"message": "添加项目成员失败: " + err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "添加项目成员成功",
	})
}

// RemoveProjectMember 移除项目成员
func (c *ProjectController) RemoveProjectMember(ctx *gin.Context) {
	projectID, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的项目ID",
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

	// 获取当前用户ID
	currentUserID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "未授权：用户未登录",
		})
		return
	}

	// 移除项目成员
	err = c.projectService.RemoveProjectMember(
		uint(projectID),
		uint(userID),
		uint(currentUserID.(uint)),
	)

	if err != nil {
		statusCode := http.StatusInternalServerError
		if err == service.ErrProjectNotFound || err == service.ErrMemberNotFound {
			statusCode = http.StatusNotFound
		} else if err == service.ErrNoProjectPermission {
			statusCode = http.StatusForbidden
		} else if err == service.ErrCannotRemoveSelf {
			statusCode = http.StatusBadRequest
		}

		ctx.JSON(statusCode, gin.H{
			"code":    statusCode,
			"message": "移除项目成员失败: " + err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "移除项目成员成功",
	})
}

// UpdateMemberRole 更新项目成员角色
func (c *ProjectController) UpdateMemberRole(ctx *gin.Context) {
	projectID, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的项目ID",
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

	// 获取当前用户ID
	currentUserID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "未授权：用户未登录",
		})
		return
	}

	// 更新项目成员角色
	err = c.projectService.UpdateProjectMemberRole(
		uint(projectID),
		uint(userID),
		req.Role,
		uint(currentUserID.(uint)),
	)

	if err != nil {
		statusCode := http.StatusInternalServerError
		if err == service.ErrProjectNotFound || err == service.ErrMemberNotFound {
			statusCode = http.StatusNotFound
		} else if err == service.ErrNoProjectPermission {
			statusCode = http.StatusForbidden
		} else if err == service.ErrInvalidRole {
			statusCode = http.StatusBadRequest
		}

		ctx.JSON(statusCode, gin.H{
			"code":    statusCode,
			"message": "更新项目成员角色失败: " + err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "更新项目成员角色成功",
	})
}
