package service

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"mime/multipart"
	"path/filepath"
	"strconv"
	"time"

	"oss-backend/internal/model/entity"
	"oss-backend/internal/repository"
	"oss-backend/pkg/minio"
)

// 错误定义
var (
	ErrFileNotFound      = errors.New("文件不存在")
	ErrFolderNotEmpty    = errors.New("文件夹不为空")
	ErrInvalidPath       = errors.New("无效的路径")
	ErrInvalidFileType   = errors.New("无效的文件类型")
	ErrFileAlreadyExists = errors.New("文件已存在")
	ErrShareExpired      = errors.New("分享已过期")
	ErrShareNotFound     = errors.New("分享不存在")
	ErrNoPermission      = errors.New("无权限操作")
	ErrInvalidFileID     = errors.New("无效的文件ID")
	ErrStorageFailed     = errors.New("存储操作失败")
)

// FileInfo 文件信息
type FileInfo struct {
	ID        uint   `json:"id"`
	Name      string `json:"name"`
	Path      string `json:"path"`
	Size      int64  `json:"size"`
	Type      string `json:"type"`
	IsDir     bool   `json:"is_dir"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	OwnerID   uint   `json:"owner_id"`
	ParentID  uint   `json:"parent_id,omitempty"`
}

// FileVersion 文件版本信息
type FileVersion struct {
	ID        uint   `json:"id"`
	FileID    uint   `json:"file_id"`
	Version   int    `json:"version"`
	Size      int64  `json:"size"`
	CreatedAt string `json:"created_at"`
	CreatorID uint   `json:"creator_id"`
}

// ShareInfo 分享信息
type ShareInfo struct {
	ID        uint   `json:"id"`
	FileID    uint   `json:"file_id"`
	Code      string `json:"code"`
	ExpiredAt string `json:"expired_at,omitempty"`
	CreatedAt string `json:"created_at"`
	CreatorID uint   `json:"creator_id"`
}

// CreateShareRequest 创建分享请求
type CreateShareRequest struct {
	FileID    uint   `json:"file_id" binding:"required"`
	ExpiredAt string `json:"expired_at,omitempty"`
}

// CreateFolderRequest 创建文件夹请求
type CreateFolderRequest struct {
	Name     string `json:"name" binding:"required"`
	ParentID uint   `json:"parent_id"`
}

// RenameFileRequest 重命名文件请求
type RenameFileRequest struct {
	Name string `json:"name" binding:"required"`
}

// MoveFileRequest 移动文件请求
type MoveFileRequest struct {
	TargetID uint `json:"target_id" binding:"required"`
}

// FileService 文件服务接口
type FileService interface {
	// 上传文件
	UploadFile(ctx context.Context, file *multipart.FileHeader, path string, userID uint) (*FileInfo, error)
	// 创建文件夹
	CreateFolder(ctx context.Context, name string, parentID uint, userID uint) (*FileInfo, error)
	// 获取文件信息
	GetFile(ctx context.Context, fileID uint, userID uint) (*FileInfo, error)
	// 列出文件
	ListFiles(ctx context.Context, parentID uint, userID uint) ([]*FileInfo, error)
	// 获取文件内容
	GetFileContent(ctx context.Context, fileID uint, userID uint) (io.ReadCloser, string, error)
	// 重命名文件
	RenameFile(ctx context.Context, fileID uint, newName string, userID uint) (*FileInfo, error)
	// 移动文件
	MoveFile(ctx context.Context, fileID uint, targetID uint, userID uint) (*FileInfo, error)
	// 复制文件
	CopyFile(ctx context.Context, fileID uint, targetID uint, userID uint) (*FileInfo, error)
	// 删除文件(逻辑删除)
	DeleteFile(ctx context.Context, fileID uint, userID uint) error
	// 恢复文件
	RestoreFile(ctx context.Context, fileID uint, userID uint) (*FileInfo, error)
	// 永久删除文件
	PermanentDeleteFile(ctx context.Context, fileID uint, userID uint) error
	// 创建分享
	CreateFileShare(ctx context.Context, fileID uint, expiredAt string, userID uint) (*ShareInfo, error)
	// 获取分享信息
	GetFileShare(ctx context.Context, code string) (*ShareInfo, *FileInfo, error)
	// 取消分享
	CancelFileShare(ctx context.Context, shareID uint, userID uint) error
	// 更新文件版本
	UpdateFileVersion(ctx context.Context, fileID uint, file *multipart.FileHeader, userID uint) (*FileVersion, error)
	// 获取文件历史版本
	GetFileVersions(ctx context.Context, fileID uint, userID uint) ([]*FileVersion, error)
	// 回滚到历史版本
	RollbackToVersion(ctx context.Context, fileID uint, versionID uint, userID uint) (*FileInfo, error)
}

// fileService 文件服务实现
type fileService struct {
	fileRepo    repository.FileRepository
	minioClient *minio.Client
}

// NewFileService 创建文件服务
func NewFileService(fileRepo repository.FileRepository, minioClient *minio.Client) FileService {
	return &fileService{
		fileRepo:    fileRepo,
		minioClient: minioClient,
	}
}

// 转换实体到FileInfo
func convertToFileInfo(file *entity.File) *FileInfo {
	return &FileInfo{
		ID:        uint(file.ID),
		Name:      file.FileName,
		Path:      file.FilePath,
		Size:      file.FileSize,
		Type:      file.MimeType,
		IsDir:     file.IsFolder,
		CreatedAt: file.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt: file.UpdatedAt.Format("2006-01-02 15:04:05"),
		OwnerID:   uint(file.UploaderID),
	}
}

// 转换实体到FileVersion
func convertToFileVersion(version *entity.FileVersion) *FileVersion {
	return &FileVersion{
		ID:        uint(version.ID),
		FileID:    uint(version.FileID),
		Version:   version.Version,
		Size:      version.FileSize,
		CreatedAt: version.CreatedAt.Format("2006-01-02 15:04:05"),
		CreatorID: uint(version.UploaderID),
	}
}

// 转换实体到ShareInfo
func convertToShareInfo(share *entity.FileShare) *ShareInfo {
	shareInfo := &ShareInfo{
		ID:        uint(share.ID),
		FileID:    uint(share.FileID),
		Code:      share.ShareCode,
		CreatedAt: share.CreatedAt.Format("2006-01-02 15:04:05"),
		CreatorID: uint(share.UserID),
	}
	if share.ExpireAt != nil {
		shareInfo.ExpiredAt = share.ExpireAt.Format("2006-01-02 15:04:05")
	}
	return shareInfo
}

// 生成随机分享码
func generateShareCode() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const length = 8

	rand.Seed(time.Now().UnixNano())
	code := make([]byte, length)
	for i := range code {
		code[i] = charset[rand.Intn(len(charset))]
	}
	return string(code)
}

// 检查用户是否有权限访问文件
func (s *fileService) checkFilePermission(ctx context.Context, fileID uint64, userID uint) (*entity.File, error) {
	file, err := s.fileRepo.GetFileByID(ctx, fileID)
	if err != nil {
		return nil, err
	}

	// 简单权限检查：文件所有者才能操作
	if file.UploaderID != uint64(userID) && userID != 0 { // userID为0表示系统操作，如分享下载
		return nil, ErrNoPermission
	}

	return file, nil
}

// UploadFile 上传文件
func (s *fileService) UploadFile(ctx context.Context, file *multipart.FileHeader, path string, userID uint) (*FileInfo, error) {
	// 打开上传的文件
	src, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer src.Close()

	// 生成文件哈希值作为存储名称
	fileHash := generateRandomHash()

	// 获取文件扩展名
	ext := filepath.Ext(file.Filename)

	// 存储到MinIO
	objectName := fmt.Sprintf("files/%s%s", fileHash, ext)
	contentType := file.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	// 上传到MinIO
	err = s.minioClient.PutObject(ctx, "oss-bucket", objectName, src, file.Size, minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		return nil, ErrStorageFailed
	}

	// 保存文件信息到数据库
	newFile := &entity.File{
		FileName:   file.Filename,
		FilePath:   path,
		FullPath:   filepath.Join(path, file.Filename),
		FileHash:   fileHash,
		FileSize:   file.Size,
		MimeType:   contentType,
		Extension:  ext,
		IsFolder:   false,
		UploaderID: uint64(userID),
	}

	if err := s.fileRepo.SaveFile(ctx, newFile); err != nil {
		return nil, err
	}

	// 创建初始版本
	version := &entity.FileVersion{
		FileID:     newFile.ID,
		Version:    1,
		FileHash:   fileHash,
		FileSize:   file.Size,
		UploaderID: uint64(userID),
	}

	if err := s.fileRepo.CreateFileVersion(ctx, version); err != nil {
		return nil, err
	}

	return convertToFileInfo(newFile), nil
}

// CreateFolder 创建文件夹
func (s *fileService) CreateFolder(ctx context.Context, name string, parentID uint, userID uint) (*FileInfo, error) {
	// 创建文件夹条目
	folder := &entity.File{
		FileName:   name,
		FilePath:   "", // 实际路径将根据parentID构建
		FullPath:   "", // 同上
		FileHash:   "",
		FileSize:   0,
		MimeType:   "application/directory",
		Extension:  "",
		IsFolder:   true,
		UploaderID: uint64(userID),
	}

	if err := s.fileRepo.SaveFile(ctx, folder); err != nil {
		return nil, err
	}

	return convertToFileInfo(folder), nil
}

// GetFile 获取文件信息
func (s *fileService) GetFile(ctx context.Context, fileID uint, userID uint) (*FileInfo, error) {
	file, err := s.checkFilePermission(ctx, uint64(fileID), userID)
	if err != nil {
		return nil, err
	}

	return convertToFileInfo(file), nil
}

// ListFiles 列出文件
func (s *fileService) ListFiles(ctx context.Context, parentID uint, userID uint) ([]*FileInfo, error) {
	// 这里简化了，实际应该添加更多权限验证逻辑
	files, err := s.fileRepo.ListFilesByParent(ctx, 0, uint64(parentID), uint64(userID))
	if err != nil {
		return nil, err
	}

	var fileInfos []*FileInfo
	for _, file := range files {
		fileInfos = append(fileInfos, convertToFileInfo(file))
	}

	return fileInfos, nil
}

// GetFileContent 获取文件内容
func (s *fileService) GetFileContent(ctx context.Context, fileID uint, userID uint) (io.ReadCloser, string, error) {
	file, err := s.checkFilePermission(ctx, uint64(fileID), userID)
	if err != nil {
		return nil, "", err
	}

	if file.IsFolder {
		return nil, "", errors.New("不能直接获取文件夹内容")
	}

	// 从MinIO获取文件
	objectName := fmt.Sprintf("files/%s%s", file.FileHash, file.Extension)
	object, err := s.minioClient.GetObject(ctx, "oss-bucket", objectName, minio.GetObjectOptions{})
	if err != nil {
		return nil, "", ErrStorageFailed
	}

	return object, file.FileName, nil
}

// RenameFile 重命名文件
func (s *fileService) RenameFile(ctx context.Context, fileID uint, newName string, userID uint) (*FileInfo, error) {
	file, err := s.checkFilePermission(ctx, uint64(fileID), userID)
	if err != nil {
		return nil, err
	}

	// 更新文件名
	file.FileName = newName
	file.FullPath = filepath.Join(filepath.Dir(file.FullPath), newName)

	if err := s.fileRepo.UpdateFile(ctx, file); err != nil {
		return nil, err
	}

	return convertToFileInfo(file), nil
}

// MoveFile 移动文件
func (s *fileService) MoveFile(ctx context.Context, fileID uint, targetID uint, userID uint) (*FileInfo, error) {
	file, err := s.checkFilePermission(ctx, uint64(fileID), userID)
	if err != nil {
		return nil, err
	}

	// 获取目标文件夹信息
	targetFolder, err := s.fileRepo.GetFileByID(ctx, uint64(targetID))
	if err != nil {
		return nil, err
	}

	if !targetFolder.IsFolder {
		return nil, errors.New("目标必须是文件夹")
	}

	// 更新文件路径
	file.FilePath = targetFolder.FullPath
	file.FullPath = filepath.Join(targetFolder.FullPath, file.FileName)

	if err := s.fileRepo.UpdateFile(ctx, file); err != nil {
		return nil, err
	}

	return convertToFileInfo(file), nil
}

// CopyFile 复制文件
func (s *fileService) CopyFile(ctx context.Context, fileID uint, targetID uint, userID uint) (*FileInfo, error) {
	originalFile, err := s.checkFilePermission(ctx, uint64(fileID), userID)
	if err != nil {
		return nil, err
	}

	// 获取目标文件夹信息
	targetFolder, err := s.fileRepo.GetFileByID(ctx, uint64(targetID))
	if err != nil {
		return nil, err
	}

	if !targetFolder.IsFolder {
		return nil, errors.New("目标必须是文件夹")
	}

	// 创建新文件记录
	newFile := &entity.File{
		FileName:       originalFile.FileName,
		FilePath:       targetFolder.FullPath,
		FullPath:       filepath.Join(targetFolder.FullPath, originalFile.FileName),
		FileHash:       originalFile.FileHash, // 复用相同的文件内容
		FileSize:       originalFile.FileSize,
		MimeType:       originalFile.MimeType,
		Extension:      originalFile.Extension,
		IsFolder:       originalFile.IsFolder,
		UploaderID:     uint64(userID),
		CurrentVersion: 1,
	}

	if err := s.fileRepo.SaveFile(ctx, newFile); err != nil {
		return nil, err
	}

	// 创建初始版本
	version := &entity.FileVersion{
		FileID:     newFile.ID,
		Version:    1,
		FileHash:   originalFile.FileHash,
		FileSize:   originalFile.FileSize,
		UploaderID: uint64(userID),
	}

	if err := s.fileRepo.CreateFileVersion(ctx, version); err != nil {
		return nil, err
	}

	return convertToFileInfo(newFile), nil
}

// DeleteFile 删除文件
func (s *fileService) DeleteFile(ctx context.Context, fileID uint, userID uint) error {
	// 检查文件是否存在并验证权限
	_, err := s.checkFilePermission(ctx, uint64(fileID), userID)
	if err != nil {
		return err
	}

	// 逻辑删除文件
	return s.fileRepo.DeleteFile(ctx, uint64(fileID), uint64(userID))
}

// RestoreFile 恢复文件
func (s *fileService) RestoreFile(ctx context.Context, fileID uint, userID uint) (*FileInfo, error) {
	// 获取文件，即使已标记为删除
	file, err := s.fileRepo.GetFileByID(ctx, uint64(fileID))
	if err != nil {
		return nil, err
	}

	// 检查权限
	if file.UploaderID != uint64(userID) {
		return nil, ErrNoPermission
	}

	// 恢复文件
	file.IsDeleted = false
	file.DeletedAt = nil
	file.DeletedBy = nil

	if err := s.fileRepo.UpdateFile(ctx, file); err != nil {
		return nil, err
	}

	return convertToFileInfo(file), nil
}

// PermanentDeleteFile 永久删除文件
func (s *fileService) PermanentDeleteFile(ctx context.Context, fileID uint, userID uint) error {
	// 检查文件是否存在并验证权限
	_, err := s.checkFilePermission(ctx, uint64(fileID), userID)
	if err != nil {
		return err
	}

	// 永久删除文件
	return s.fileRepo.PermanentDeleteFile(ctx, uint64(fileID))

	// 注意：这里可以添加代码来从MinIO删除文件，但通常会保留文件内容以防万一
}

// CreateFileShare 创建分享
func (s *fileService) CreateFileShare(ctx context.Context, fileID uint, expiredAt string, userID uint) (*ShareInfo, error) {
	// 检查文件是否存在并验证权限
	_, err := s.checkFilePermission(ctx, uint64(fileID), userID)
	if err != nil {
		return nil, err
	}

	// 生成分享码
	shareCode := generateShareCode()

	// 创建分享记录
	share := &entity.FileShare{
		FileID:    uint64(fileID),
		UserID:    uint64(userID),
		ShareCode: shareCode,
	}

	// 设置过期时间
	if expiredAt != "" {
		expTime, err := time.Parse("2006-01-02 15:04:05", expiredAt)
		if err != nil {
			return nil, errors.New("无效的过期时间格式")
		}
		share.ExpireAt = &expTime
	}

	// 保存分享记录
	if err := s.fileRepo.CreateFileShare(ctx, share); err != nil {
		return nil, err
	}

	return convertToShareInfo(share), nil
}

// GetFileShare 获取分享信息
func (s *fileService) GetFileShare(ctx context.Context, code string) (*ShareInfo, *FileInfo, error) {
	// 获取分享记录
	share, err := s.fileRepo.GetFileShareByCode(ctx, code)
	if err != nil {
		if err == repository.ErrShareNotFound {
			return nil, nil, ErrShareNotFound
		}
		return nil, nil, err
	}

	// 检查是否过期
	if share.ExpireAt != nil && share.ExpireAt.Before(time.Now()) {
		return nil, nil, ErrShareExpired
	}

	// 获取文件信息
	file, err := s.fileRepo.GetFileByID(ctx, share.FileID)
	if err != nil {
		return nil, nil, err
	}

	return convertToShareInfo(share), convertToFileInfo(file), nil
}

// CancelFileShare 取消分享
func (s *fileService) CancelFileShare(ctx context.Context, shareID uint, userID uint) error {
	// 从数据库获取分享记录
	// 使用ID查询而非ShareCode
	share, err := s.fileRepo.GetFileShareByCode(ctx, strconv.FormatUint(uint64(shareID), 10))
	if err != nil {
		if err == repository.ErrShareNotFound {
			return ErrShareNotFound
		}
		return err
	}

	// 检查权限
	if share.UserID != uint64(userID) {
		return ErrNoPermission
	}

	// 删除分享
	return s.fileRepo.DeleteFileShare(ctx, share.ID)
}

// UpdateFileVersion 更新文件版本
func (s *fileService) UpdateFileVersion(ctx context.Context, fileID uint, file *multipart.FileHeader, userID uint) (*FileVersion, error) {
	// 检查文件是否存在
	existingFile, err := s.checkFilePermission(ctx, uint64(fileID), userID)
	if err != nil {
		return nil, err
	}

	if existingFile.IsFolder {
		return nil, errors.New("不能更新文件夹的版本")
	}

	// 打开上传的文件
	src, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer src.Close()

	// 生成新文件哈希值
	fileHash := generateRandomHash()

	// 存储到MinIO
	objectName := fmt.Sprintf("files/%s%s", fileHash, existingFile.Extension)
	contentType := file.Header.Get("Content-Type")
	if contentType == "" {
		contentType = existingFile.MimeType
	}

	// 上传到MinIO
	err = s.minioClient.PutObject(ctx, "oss-bucket", objectName, src, file.Size, minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		return nil, ErrStorageFailed
	}

	// 获取当前版本
	versions, err := s.fileRepo.GetFileVersions(ctx, uint64(fileID))
	if err != nil {
		return nil, err
	}

	newVersion := 1
	if len(versions) > 0 {
		newVersion = versions[0].Version + 1
	}

	// 创建新版本
	fileVersion := &entity.FileVersion{
		FileID:     uint64(fileID),
		Version:    newVersion,
		FileHash:   fileHash,
		FileSize:   file.Size,
		UploaderID: uint64(userID),
	}

	if err := s.fileRepo.CreateFileVersion(ctx, fileVersion); err != nil {
		return nil, err
	}

	// 更新文件信息
	existingFile.FileHash = fileHash
	existingFile.FileSize = file.Size
	existingFile.MimeType = contentType
	existingFile.CurrentVersion = newVersion

	if err := s.fileRepo.UpdateFile(ctx, existingFile); err != nil {
		return nil, err
	}

	return convertToFileVersion(fileVersion), nil
}

// GetFileVersions 获取文件历史版本
func (s *fileService) GetFileVersions(ctx context.Context, fileID uint, userID uint) ([]*FileVersion, error) {
	// 检查文件是否存在
	_, err := s.checkFilePermission(ctx, uint64(fileID), userID)
	if err != nil {
		return nil, err
	}

	// 获取版本列表
	versions, err := s.fileRepo.GetFileVersions(ctx, uint64(fileID))
	if err != nil {
		return nil, err
	}

	var versionInfos []*FileVersion
	for _, version := range versions {
		versionInfos = append(versionInfos, convertToFileVersion(version))
	}

	return versionInfos, nil
}

// RollbackToVersion 回滚到历史版本
func (s *fileService) RollbackToVersion(ctx context.Context, fileID uint, versionID uint, userID uint) (*FileInfo, error) {
	// 检查文件是否存在
	file, err := s.checkFilePermission(ctx, uint64(fileID), userID)
	if err != nil {
		return nil, err
	}

	// 获取版本信息
	version, err := s.fileRepo.GetFileVersion(ctx, uint64(fileID), int(versionID))
	if err != nil {
		return nil, err
	}

	// 更新文件信息
	file.FileHash = version.FileHash
	file.FileSize = version.FileSize
	file.CurrentVersion = version.Version

	if err := s.fileRepo.UpdateFile(ctx, file); err != nil {
		return nil, err
	}

	return convertToFileInfo(file), nil
}

// 生成随机哈希值
func generateRandomHash() string {
	// 初始化随机数生成器
	rand.Seed(time.Now().UnixNano())

	// 生成随机字节
	randomBytes := make([]byte, 16)
	for i := range randomBytes {
		randomBytes[i] = byte(rand.Intn(256))
	}

	// 计算MD5哈希
	hasher := md5.New()
	hasher.Write(randomBytes)
	return hex.EncodeToString(hasher.Sum(nil))
}
