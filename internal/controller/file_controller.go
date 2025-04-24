package controller

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"oss-backend/internal/model/dto"
	"oss-backend/internal/model/entity"
	"oss-backend/internal/service"
	"oss-backend/pkg/common"
)

// FileController 文件控制器
type FileController struct {
	fileService    service.FileService
	projectService service.ProjectService
	authService    service.AuthService
}

// NewFileController 创建文件控制器
func NewFileController(fileService service.FileService, projectService service.ProjectService, authService service.AuthService) *FileController {
	return &FileController{
		fileService:    fileService,
		projectService: projectService,
		authService:    authService,
	}
}

// Upload 上传文件
// @Summary 上传文件
// @Description 上传文件到指定项目和路径
// @Tags 文件管理
// @Accept multipart/form-data
// @Produce json
// @Param Authorization header string true "Bearer {{token}}"
// @Param project_id formData int true "项目ID"
// @Param path formData string false "上传路径，默认为根目录"
// @Param comment formData string false "文件注释"
// @Param overwrite formData bool false "是否覆盖同名文件"
// @Param file formData file true "上传的文件"
// @Success 200 {object} common.Response{data=dto.FileResponse} "成功"
// @Failure 400 {object} common.Response "请求参数错误"
// @Failure 401 {object} common.Response "未授权"
// @Failure 403 {object} common.Response "权限不足"
// @Failure 500 {object} common.Response "内部服务器错误"
// @Router /api/oss/file/upload [post]
func (c *FileController) Upload(ctx *gin.Context) {
	// 获取当前用户ID
	userIDValue, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, common.ErrorResponse("未授权"))
		return
	}
	userID := userIDValue.(uint64)

	// 绑定请求参数
	var req dto.FileUploadRequest
	if err := ctx.ShouldBind(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, common.ErrorResponse(err.Error()))
		return
	}

	// 获取上传文件
	file, err := ctx.FormFile("file")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, common.ErrorResponse("获取上传文件失败: "+err.Error()))
		return
	}

	// 检查项目权限 (需要写入权限)
	canWrite, err := c.authService.CheckUserProjectPermission(ctx, userID, req.ProjectID, []string{"admin", "editor"})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, common.ErrorResponse("检查权限失败: "+err.Error()))
		return
	}
	if !canWrite {
		ctx.JSON(http.StatusForbidden, common.ErrorResponse("没有项目写入权限"))
		return
	}

	// 上传文件
	uploadedFile, err := c.fileService.Upload(ctx, req.ProjectID, userID, file, req.Path)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, common.ErrorResponse("上传文件失败: "+err.Error()))
		return
	}

	// 构建响应
	response := buildFileResponse(uploadedFile)

	ctx.JSON(http.StatusOK, common.SuccessResponse(response))
}

// Download 下载文件
// @Summary 下载文件
// @Description 下载指定ID的文件
// @Tags 文件管理
// @Produce octet-stream
// @Param Authorization header string true "Bearer {{token}}"
// @Param id path int true "文件ID"
// @Success 200 {file} octet-stream "文件内容"
// @Failure 400 {object} common.Response "请求参数错误"
// @Failure 401 {object} common.Response "未授权"
// @Failure 403 {object} common.Response "权限不足"
// @Failure 404 {object} common.Response "文件不存在"
// @Failure 500 {object} common.Response "内部服务器错误"
// @Router /api/oss/file/download/{id} [get]
func (c *FileController) Download(ctx *gin.Context) {
	// 获取当前用户ID
	userIDValue, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, common.ErrorResponse("未授权"))
		return
	}
	userID := userIDValue.(uint64)

	// 获取文件ID
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, common.ErrorResponse("无效的文件ID"))
		return
	}

	// 获取文件信息
	fileInfo, err := c.fileService.GetFileInfo(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, common.ErrorResponse("获取文件信息失败: "+err.Error()))
		return
	}
	if fileInfo == nil {
		ctx.JSON(http.StatusNotFound, common.ErrorResponse("文件不存在"))
		return
	}

	// 检查项目权限 (需要读取权限)
	canRead, err := c.authService.CheckUserProjectPermission(ctx, userID, fileInfo.ProjectID, []string{"admin", "editor", "viewer"})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, common.ErrorResponse("检查权限失败: "+err.Error()))
		return
	}
	if !canRead {
		ctx.JSON(http.StatusForbidden, common.ErrorResponse("没有文件读取权限"))
		return
	}

	// 下载文件
	fileReader, file, err := c.fileService.Download(ctx, id, userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, common.ErrorResponse("下载文件失败: "+err.Error()))
		return
	}
	defer fileReader.Close()

	// 设置响应头
	ctx.Header("Content-Description", "File Transfer")
	ctx.Header("Content-Transfer-Encoding", "binary")
	ctx.Header("Content-Disposition", "attachment; filename="+file.FileName)
	ctx.Header("Content-Type", file.MimeType)
	ctx.Header("Content-Length", strconv.FormatInt(file.FileSize, 10))
	ctx.Header("Accept-Ranges", "bytes")

	// 发送文件内容
	ctx.DataFromReader(http.StatusOK, file.FileSize, file.MimeType, fileReader, nil)
}

// ListFiles 获取文件列表
// @Summary 获取文件列表
// @Description 获取指定项目和路径下的文件列表
// @Tags 文件管理
// @Produce json
// @Param Authorization header string true "Bearer {{token}}"
// @Param project_id query int true "项目ID"
// @Param path query string false "文件路径，默认为根目录"
// @Param recursive query bool false "是否递归获取子目录"
// @Param page query int false "页码，默认1"
// @Param size query int false "每页大小，默认20"
// @Success 200 {object} common.Response{data=dto.FileListResponse} "成功"
// @Failure 400 {object} common.Response "请求参数错误"
// @Failure 401 {object} common.Response "未授权"
// @Failure 403 {object} common.Response "权限不足"
// @Failure 500 {object} common.Response "内部服务器错误"
// @Router /api/oss/file/list [get]
func (c *FileController) ListFiles(ctx *gin.Context) {
	// 获取当前用户ID
	userIDValue, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, common.ErrorResponse("未授权"))
		return
	}
	userID := userIDValue.(uint64)

	// 绑定请求参数
	var req dto.FileListRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, common.ErrorResponse(err.Error()))
		return
	}

	// 检查项目权限 (需要读取权限)
	canRead, err := c.authService.CheckUserProjectPermission(ctx, userID, req.ProjectID, []string{"admin", "editor", "viewer"})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, common.ErrorResponse("检查权限失败: "+err.Error()))
		return
	}
	if !canRead {
		ctx.JSON(http.StatusForbidden, common.ErrorResponse("没有项目读取权限"))
		return
	}

	// 获取文件列表
	files, total, err := c.fileService.ListFiles(ctx, req.ProjectID, req.Path, req.Recursive, req.Page, req.Size)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, common.ErrorResponse("获取文件列表失败: "+err.Error()))
		return
	}

	// 构建响应
	response := dto.FileListResponse{
		Total: total,
		Items: make([]dto.FileResponse, 0, len(files)),
	}

	for _, file := range files {
		response.Items = append(response.Items, buildFileResponse(file))
	}

	ctx.JSON(http.StatusOK, common.SuccessResponse(response))
}

// CreateFolder 创建文件夹
// @Summary 创建文件夹
// @Description 在指定项目和路径下创建文件夹
// @Tags 文件管理
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer {{token}}"
// @Param request body dto.FileFolderCreateRequest true "创建文件夹请求"
// @Success 200 {object} common.Response{data=dto.FileResponse} "成功"
// @Failure 400 {object} common.Response "请求参数错误"
// @Failure 401 {object} common.Response "未授权"
// @Failure 403 {object} common.Response "权限不足"
// @Failure 500 {object} common.Response "内部服务器错误"
// @Router /api/oss/file/folder [post]
func (c *FileController) CreateFolder(ctx *gin.Context) {
	// 获取当前用户ID
	userIDValue, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, common.ErrorResponse("未授权"))
		return
	}
	userID := userIDValue.(uint64)

	// 绑定请求参数
	var req dto.FileFolderCreateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, common.ErrorResponse(err.Error()))
		return
	}

	// 检查项目权限 (需要写入权限)
	canWrite, err := c.authService.CheckUserProjectPermission(ctx, userID, req.ProjectID, []string{"admin", "editor"})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, common.ErrorResponse("检查权限失败: "+err.Error()))
		return
	}
	if !canWrite {
		ctx.JSON(http.StatusForbidden, common.ErrorResponse("没有项目写入权限"))
		return
	}

	// 创建文件夹
	folder, err := c.fileService.CreateFolder(ctx, req.ProjectID, userID, req.Path, req.FolderName)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, common.ErrorResponse("创建文件夹失败: "+err.Error()))
		return
	}

	// 构建响应
	response := buildFileResponse(folder)

	ctx.JSON(http.StatusOK, common.SuccessResponse(response))
}

// DeleteFile 删除文件
// @Summary 删除文件
// @Description 删除指定ID的文件或文件夹(软删除)
// @Tags 文件管理
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer {{token}}"
// @Param id path int true "文件ID"
// @Success 200 {object} common.Response "成功"
// @Failure 400 {object} common.Response "请求参数错误"
// @Failure 401 {object} common.Response "未授权"
// @Failure 403 {object} common.Response "权限不足"
// @Failure 404 {object} common.Response "文件不存在"
// @Failure 500 {object} common.Response "内部服务器错误"
// @Router /api/oss/file/delete/{id} [get]
func (c *FileController) DeleteFile(ctx *gin.Context) {
	// 获取当前用户ID
	userIDValue, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, common.ErrorResponse("未授权"))
		return
	}
	userID := userIDValue.(uint64)

	// 获取文件ID
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, common.ErrorResponse("无效的文件ID"))
		return
	}

	// 获取文件信息
	fileInfo, err := c.fileService.GetFileInfo(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, common.ErrorResponse("获取文件信息失败: "+err.Error()))
		return
	}
	if fileInfo == nil {
		ctx.JSON(http.StatusNotFound, common.ErrorResponse("文件不存在"))
		return
	}

	// 检查项目权限 (需要写入权限)
	canWrite, err := c.authService.CheckUserProjectPermission(ctx, userID, fileInfo.ProjectID, []string{"admin", "editor"})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, common.ErrorResponse("检查权限失败: "+err.Error()))
		return
	}

	// 检查权限：管理员可以删除任何文件，编辑者只能删除自己的文件
	isAdmin, err := c.authService.IsProjectAdmin(ctx, userID, fileInfo.ProjectID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, common.ErrorResponse("检查管理员权限失败: "+err.Error()))
		return
	}

	if !canWrite || (fileInfo.UploaderID != userID && !isAdmin) {
		ctx.JSON(http.StatusForbidden, common.ErrorResponse("没有删除此文件的权限"))
		return
	}

	// 删除文件
	err = c.fileService.DeleteFile(ctx, id, userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, common.ErrorResponse("删除文件失败: "+err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, common.SuccessResponse(nil))
}

// GetFileVersions 获取文件版本列表
// @Summary 获取文件版本列表
// @Description 获取指定文件的所有版本信息
// @Tags 文件管理
// @Produce json
// @Param Authorization header string true "Bearer {{token}}"
// @Param id path int true "文件ID"
// @Success 200 {object} common.Response{data=dto.FileVersionListResponse} "成功"
// @Failure 400 {object} common.Response "请求参数错误"
// @Failure 401 {object} common.Response "未授权"
// @Failure 403 {object} common.Response "权限不足"
// @Failure 404 {object} common.Response "文件不存在"
// @Failure 500 {object} common.Response "内部服务器错误"
// @Router /api/oss/file/versions/{id} [get]
func (c *FileController) GetFileVersions(ctx *gin.Context) {
	// 获取当前用户ID
	userIDValue, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, common.ErrorResponse("未授权"))
		return
	}
	userID := userIDValue.(uint64)

	// 获取文件ID
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, common.ErrorResponse("无效的文件ID"))
		return
	}

	// 获取文件信息
	fileInfo, err := c.fileService.GetFileInfo(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, common.ErrorResponse("获取文件信息失败: "+err.Error()))
		return
	}
	if fileInfo == nil {
		ctx.JSON(http.StatusNotFound, common.ErrorResponse("文件不存在"))
		return
	}

	// 检查项目权限 (需要读取权限)
	canRead, err := c.authService.CheckUserProjectPermission(ctx, userID, fileInfo.ProjectID, []string{"admin", "editor", "viewer"})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, common.ErrorResponse("检查权限失败: "+err.Error()))
		return
	}
	if !canRead {
		ctx.JSON(http.StatusForbidden, common.ErrorResponse("没有查看此文件的权限"))
		return
	}

	// 获取文件版本列表
	versions, err := c.fileService.GetFileVersions(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, common.ErrorResponse("获取文件版本失败: "+err.Error()))
		return
	}

	// 构建响应
	response := dto.FileVersionListResponse{
		FileID: id,
		Total:  len(versions),
		Items:  make([]dto.FileVersionResponse, 0, len(versions)),
	}

	for _, version := range versions {
		item := dto.FileVersionResponse{
			ID:           version.ID,
			FileID:       version.FileID,
			Version:      version.Version,
			FileHash:     version.FileHash,
			FileSize:     version.FileSize,
			UploaderID:   version.UploaderID,
			UploaderName: version.Uploader.Name,
			CreatedAt:    version.CreatedAt,
			Comment:      version.Comment,
		}
		response.Items = append(response.Items, item)
	}

	ctx.JSON(http.StatusOK, common.SuccessResponse(response))
}

// CreateShare 创建文件分享
// @Summary 创建文件分享
// @Description 创建文件分享链接
// @Tags 文件分享
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer {{token}}"
// @Param request body dto.FileShareCreateRequest true "创建分享请求"
// @Success 200 {object} common.Response{data=dto.FileShareResponse} "成功"
// @Failure 400 {object} common.Response "请求参数错误"
// @Failure 401 {object} common.Response "未授权"
// @Failure 403 {object} common.Response "权限不足"
// @Failure 404 {object} common.Response "文件不存在"
// @Failure 500 {object} common.Response "内部服务器错误"
// @Router /api/oss/file/share [post]
func (c *FileController) CreateShare(ctx *gin.Context) {
	// 获取当前用户ID
	userIDValue, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, common.ErrorResponse("未授权"))
		return
	}
	userID := userIDValue.(uint64)

	// 绑定请求参数
	var req dto.FileShareCreateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, common.ErrorResponse(err.Error()))
		return
	}

	// 获取文件信息
	fileInfo, err := c.fileService.GetFileInfo(ctx, req.FileID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, common.ErrorResponse("获取文件信息失败: "+err.Error()))
		return
	}
	if fileInfo == nil {
		ctx.JSON(http.StatusNotFound, common.ErrorResponse("文件不存在"))
		return
	}

	// 检查项目权限 (需要读取权限)
	canRead, err := c.authService.CheckUserProjectPermission(ctx, userID, fileInfo.ProjectID, []string{"admin", "editor", "viewer"})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, common.ErrorResponse("检查权限失败: "+err.Error()))
		return
	}
	if !canRead {
		ctx.JSON(http.StatusForbidden, common.ErrorResponse("没有分享此文件的权限"))
		return
	}

	// 创建分享
	share, err := c.fileService.CreateShare(ctx, req.FileID, userID, req.Password, req.ExpireHours, req.DownloadLimit)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, common.ErrorResponse("创建分享失败: "+err.Error()))
		return
	}

	// 构建响应
	response := dto.FileShareResponse{
		ID:            share.ID,
		FileID:        share.FileID,
		FileName:      share.File.FileName,
		FileSize:      share.File.FileSize,
		MimeType:      share.File.MimeType,
		ShareCode:     share.ShareCode,
		HasPassword:   share.Password != "",
		ExpireAt:      share.ExpireAt,
		DownloadLimit: share.DownloadLimit,
		DownloadCount: share.DownloadCount,
		CreatedAt:     share.CreatedAt,
		CreatorName:   share.User.Name,
	}

	ctx.JSON(http.StatusOK, common.SuccessResponse(response))
}

// GetShareInfo 获取分享信息
// @Summary 获取分享信息
// @Description 根据分享码获取分享信息
// @Tags 文件分享
// @Produce json
// @Param code path string true "分享码"
// @Success 200 {object} common.Response{data=dto.FileShareResponse} "成功"
// @Failure 400 {object} common.Response "请求参数错误"
// @Failure 404 {object} common.Response "分享不存在"
// @Failure 500 {object} common.Response "内部服务器错误"
// @Router /api/oss/share/{code} [get]
func (c *FileController) GetShareInfo(ctx *gin.Context) {
	// 获取分享码
	code := ctx.Param("code")
	if code == "" {
		ctx.JSON(http.StatusBadRequest, common.ErrorResponse("分享码不能为空"))
		return
	}

	// 获取分享信息
	share, err := c.fileService.GetShareInfo(ctx, code)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, common.ErrorResponse("获取分享信息失败: "+err.Error()))
		return
	}
	if share == nil {
		ctx.JSON(http.StatusNotFound, common.ErrorResponse("分享不存在或已过期"))
		return
	}

	// 构建响应
	response := dto.FileShareResponse{
		ID:            share.ID,
		FileID:        share.FileID,
		FileName:      share.File.FileName,
		FileSize:      share.File.FileSize,
		MimeType:      share.File.MimeType,
		ShareCode:     share.ShareCode,
		HasPassword:   share.Password != "",
		ExpireAt:      share.ExpireAt,
		DownloadLimit: share.DownloadLimit,
		DownloadCount: share.DownloadCount,
		CreatedAt:     share.CreatedAt,
		CreatorName:   share.User.Name,
	}

	ctx.JSON(http.StatusOK, common.SuccessResponse(response))
}

// DownloadSharedFile 下载分享文件
// @Summary 下载分享文件
// @Description 下载通过分享链接的文件
// @Tags 文件分享
// @Accept json
// @Produce octet-stream
// @Param request body dto.FileShareAccessRequest true "访问分享请求"
// @Success 200 {file} octet-stream "文件内容"
// @Failure 400 {object} common.Response "请求参数错误"
// @Failure 401 {object} common.Response "密码错误"
// @Failure 404 {object} common.Response "分享不存在"
// @Failure 500 {object} common.Response "内部服务器错误"
// @Router /api/oss/share/download [post]
func (c *FileController) DownloadSharedFile(ctx *gin.Context) {
	// 绑定请求参数
	var req dto.FileShareAccessRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, common.ErrorResponse(err.Error()))
		return
	}

	// 下载分享文件
	fileReader, file, err := c.fileService.DownloadSharedFile(ctx, req.ShareCode, req.Password)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, common.ErrorResponse("下载文件失败: "+err.Error()))
		return
	}
	defer fileReader.Close()

	// 设置响应头
	ctx.Header("Content-Description", "File Transfer")
	ctx.Header("Content-Transfer-Encoding", "binary")
	ctx.Header("Content-Disposition", "attachment; filename="+file.FileName)
	ctx.Header("Content-Type", file.MimeType)
	ctx.Header("Content-Length", strconv.FormatInt(file.FileSize, 10))
	ctx.Header("Accept-Ranges", "bytes")

	// 发送文件内容
	ctx.DataFromReader(http.StatusOK, file.FileSize, file.MimeType, fileReader, nil)
}

// 构建文件响应对象
func buildFileResponse(file *entity.File) dto.FileResponse {
	response := dto.FileResponse{
		ID:             file.ID,
		ProjectID:      file.ProjectID,
		FileName:       file.FileName,
		FilePath:       file.FilePath,
		FullPath:       file.FullPath,
		FileSize:       file.FileSize,
		MimeType:       file.MimeType,
		Extension:      file.Extension,
		IsFolder:       file.IsFolder,
		IsDeleted:      file.IsDeleted,
		UploaderID:     file.UploaderID,
		CreatedAt:      file.CreatedAt,
		UpdatedAt:      file.UpdatedAt,
		DeletedAt:      file.DeletedAt,
		DeletedBy:      file.DeletedBy,
		CurrentVersion: file.CurrentVersion,
		PreviewURL:     file.PreviewURL,
	}

	if file.Uploader.ID > 0 {
		response.UploaderName = file.Uploader.Name
	}

	if file.DeletedBy != nil && file.Deleter != nil {
		response.DeleterName = file.Deleter.Name
	}

	return response
}
