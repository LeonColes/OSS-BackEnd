package controller

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"oss-backend/internal/service"

	"github.com/gin-gonic/gin"
)

// FileController 处理文件相关请求
type FileController struct {
	fileService service.FileService
}

// NewFileController 创建新的文件控制器
func NewFileController(fileService service.FileService) *FileController {
	return &FileController{
		fileService: fileService,
	}
}

// UploadFile 上传文件
// @Summary 上传文件
// @Description 上传文件到指定文件夹
// @Tags file
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param file formData file true "文件"
// @Param folder_id formData string false "文件夹ID，不传则上传到根目录"
// @Param description formData string false "文件描述"
// @Success 200 {object} map[string]interface{} "上传成功"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /file/upload [post]
func (c *FileController) UploadFile(ctx *gin.Context) {
	userID := ctx.GetUint64("userID")
	if userID == 0 {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	file, err := ctx.FormFile("file")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无法获取上传文件"})
		return
	}

	parentIDStr := ctx.DefaultPostForm("parentID", "0")
	_, err = strconv.ParseUint(parentIDStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的父文件夹ID"})
		return
	}

	fileInfo, err := c.fileService.UploadFile(ctx, file, "", uint(userID))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("上传文件失败: %v", err)})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": fileInfo})
}

// CreateFolder 创建文件夹
// @Summary 创建文件夹
// @Description 创建新的文件夹
// @Tags file
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateFolderRequest true "文件夹信息"
// @Success 200 {object} map[string]interface{} "创建成功"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /file/folder [post]
func (c *FileController) CreateFolder(ctx *gin.Context) {
	userID := ctx.GetUint64("userID")
	if userID == 0 {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	var req CreateFolderRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求参数"})
		return
	}

	folderInfo, err := c.fileService.CreateFolder(ctx, req.Name, req.ParentID, uint(userID))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("创建文件夹失败: %v", err)})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": folderInfo})
}

// GetFile 获取文件信息
// @Summary 获取文件信息
// @Description 获取指定文件的详细信息
// @Tags file
// @Produce json
// @Security BearerAuth
// @Param id path int true "文件ID"
// @Success 200 {object} map[string]interface{} "获取成功"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Failure 403 {object} map[string]interface{} "无权访问"
// @Failure 404 {object} map[string]interface{} "文件不存在"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /file/{id} [get]
func (c *FileController) GetFile(ctx *gin.Context) {
	userID := ctx.GetUint64("userID")
	if userID == 0 {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	fileIDStr := ctx.Param("id")
	fileID, err := strconv.ParseUint(fileIDStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的文件ID"})
		return
	}

	fileInfo, err := c.fileService.GetFile(ctx, uint(fileID), uint(userID))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("获取文件信息失败: %v", err)})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": fileInfo})
}

// ListFiles 获取文件列表
// @Summary 获取文件列表
// @Description 获取指定文件夹下的文件和子文件夹
// @Tags file
// @Produce json
// @Security BearerAuth
// @Param folder_id query int false "文件夹ID，不传则获取根目录"
// @Param page query int false "页码，默认1"
// @Param size query int false "每页数量，默认20"
// @Success 200 {object} map[string]interface{} "获取成功"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Failure 403 {object} map[string]interface{} "无权访问"
// @Failure 404 {object} map[string]interface{} "文件夹不存在"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /file/list [get]
func (c *FileController) ListFiles(ctx *gin.Context) {
	userID := ctx.GetUint64("userID")
	if userID == 0 {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	parentIDStr := ctx.DefaultQuery("parentID", "0")
	_, err := strconv.ParseUint(parentIDStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的父文件夹ID"})
		return
	}

	fileList, err := c.fileService.ListFiles(ctx, 0, uint(userID))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("获取文件列表失败: %v", err)})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": fileList})
}

// DownloadFile 下载文件
// @Summary 下载文件
// @Description 下载指定文件
// @Tags file
// @Produce octet-stream
// @Security BearerAuth
// @Param id path int true "文件ID"
// @Param version query int false "文件版本号，不传则下载最新版本"
// @Success 200 {file} binary "文件内容"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Failure 403 {object} map[string]interface{} "无权访问"
// @Failure 404 {object} map[string]interface{} "文件不存在"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /file/{id}/download [get]
func (c *FileController) DownloadFile(ctx *gin.Context) {
	userID := ctx.GetUint64("userID")
	if userID == 0 {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	fileIDStr := ctx.Param("id")
	fileID, err := strconv.ParseUint(fileIDStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的文件ID"})
		return
	}

	reader, fileName, err := c.fileService.GetFileContent(ctx, uint(fileID), uint(userID))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("获取下载信息失败: %v", err)})
		return
	}

	ctx.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileName))
	ctx.Header("Content-Type", "application/octet-stream")

	ctx.DataFromReader(http.StatusOK, -1, "application/octet-stream", reader, nil)
}

// RenameFile 重命名文件
// @Summary 重命名文件
// @Description 重命名指定文件或文件夹
// @Tags file
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body RenameFileRequest true "重命名信息"
// @Success 200 {object} map[string]interface{} "重命名成功"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Failure 403 {object} map[string]interface{} "无权操作"
// @Failure 404 {object} map[string]interface{} "文件不存在"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /file/{id}/rename [post]
func (c *FileController) RenameFile(ctx *gin.Context) {
	userID := ctx.GetUint64("userID")
	if userID == 0 {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	fileIDStr := ctx.Param("id")
	fileID, err := strconv.ParseUint(fileIDStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的文件ID"})
		return
	}

	var req RenameFileRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求参数"})
		return
	}

	fileInfo, err := c.fileService.RenameFile(ctx, uint(fileID), req.Name, uint(userID))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("重命名文件失败: %v", err)})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": fileInfo})
}

// MoveFile 移动文件
// @Summary 移动文件
// @Description 移动指定文件或文件夹到新的位置
// @Tags file
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "文件ID"
// @Param request body MoveFileRequest true "移动信息"
// @Success 200 {object} map[string]interface{} "移动成功"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Failure 403 {object} map[string]interface{} "无权操作"
// @Failure 404 {object} map[string]interface{} "文件不存在"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /file/{id}/move [post]
func (c *FileController) MoveFile(ctx *gin.Context) {
	userID := ctx.GetUint64("userID")
	if userID == 0 {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	fileIDStr := ctx.Param("id")
	fileID, err := strconv.ParseUint(fileIDStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的文件ID"})
		return
	}

	var req MoveFileRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求参数"})
		return
	}

	fileInfo, err := c.fileService.MoveFile(ctx, uint(fileID), req.TargetID, uint(userID))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("移动文件失败: %v", err)})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": fileInfo})
}

// DeleteFile 删除文件
// @Summary 删除文件
// @Description 删除指定文件或文件夹（移入回收站）
// @Tags file
// @Produce json
// @Security BearerAuth
// @Param id path int true "文件ID"
// @Success 200 {object} map[string]interface{} "删除成功"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Failure 403 {object} map[string]interface{} "无权操作"
// @Failure 404 {object} map[string]interface{} "文件不存在"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /file/{id} [delete]
func (c *FileController) DeleteFile(ctx *gin.Context) {
	userID := ctx.GetUint64("userID")
	if userID == 0 {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	fileIDStr := ctx.Param("id")
	fileID, err := strconv.ParseUint(fileIDStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的文件ID"})
		return
	}

	err = c.fileService.DeleteFile(ctx, uint(fileID), uint(userID))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("删除文件失败: %v", err)})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "文件已删除"})
}

// CreateShare 创建分享链接
// @Summary 创建分享链接
// @Description 为指定文件或文件夹创建分享链接
// @Tags file
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateShareRequest true "分享请求"
// @Success 200 {object} map[string]interface{} "创建成功"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Failure 403 {object} map[string]interface{} "无权操作"
// @Failure 404 {object} map[string]interface{} "文件不存在"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /file/share [post]
func (c *FileController) CreateShare(ctx *gin.Context) {
	userID := ctx.GetUint64("userID")
	if userID == 0 {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	var req CreateShareRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求参数"})
		return
	}

	shareInfo, err := c.fileService.CreateFileShare(ctx, req.FileID, req.ExpiredAt, uint(userID))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("创建分享失败: %v", err)})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": shareInfo})
}

// GetShareFile 获取分享文件
// @Summary 获取分享文件
// @Description 获取通过分享链接访问的文件
// @Tags share
// @Accept json
// @Produce json
// @Param code path string true "分享码"
// @Param password query string false "分享密码"
// @Success 200 {object} map[string]interface{} "获取成功"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 401 {object} map[string]interface{} "密码错误"
// @Failure 404 {object} map[string]interface{} "分享不存在或已过期"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /share/{code} [get]
func (c *FileController) GetShareFile(ctx *gin.Context) {
	shareCode := ctx.Param("code")
	if shareCode == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "分享码不能为空"})
		return
	}

	shareInfo, fileInfo, err := c.fileService.GetFileShare(ctx, shareCode)
	if err != nil {
		if errors.Is(err, service.ErrShareExpired) {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "分享已过期"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("获取分享文件信息失败: %v", err)})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": fileInfo, "share": shareInfo})
}

// DownloadShareFile 下载分享文件
// @Summary 下载分享文件
// @Description 下载通过分享链接访问的文件
// @Tags share
// @Produce octet-stream
// @Param code path string true "分享码"
// @Param password query string false "分享密码"
// @Success 200 {file} binary "文件内容"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 401 {object} map[string]interface{} "密码错误"
// @Failure 404 {object} map[string]interface{} "分享不存在或已过期"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /share/{code}/download [get]
func (c *FileController) DownloadShareFile(ctx *gin.Context) {
	shareCode := ctx.Param("code")
	if shareCode == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "分享码不能为空"})
		return
	}

	reader, fileName, err := c.fileService.GetFileContent(ctx, 0, 0)
	if err != nil {
		if errors.Is(err, service.ErrShareExpired) {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "分享已过期"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("获取分享下载信息失败: %v", err)})
		return
	}

	ctx.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileName))
	ctx.Header("Content-Type", "application/octet-stream")

	ctx.DataFromReader(http.StatusOK, -1, "application/octet-stream", reader, nil)
}

// 请求结构体定义

// CreateFolderRequest 创建文件夹请求结构
type CreateFolderRequest struct {
	Name     string `json:"name" binding:"required"`
	ParentID uint   `json:"parent_id"`
}

// RenameFileRequest 重命名文件请求结构
type RenameFileRequest struct {
	Name string `json:"name" binding:"required"`
}

// MoveFileRequest 移动文件请求结构
type MoveFileRequest struct {
	TargetID uint `json:"target_id" binding:"required"`
}

// CreateShareRequest 创建分享请求结构
type CreateShareRequest struct {
	FileID    uint   `json:"file_id" binding:"required"`
	ExpiredAt string `json:"expired_at,omitempty"`
}
