package service

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"oss-backend/internal/model/entity"
	"oss-backend/internal/repository"
	"oss-backend/pkg/minio"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// FileService 文件服务接口
type FileService interface {
	// 文件操作
	Upload(ctx context.Context, projectID, uploaderID uint64, file *multipart.FileHeader, path string) (*entity.File, error)
	Download(ctx context.Context, fileID uint64, userID uint64) (io.ReadCloser, *entity.File, error)
	ListFiles(ctx context.Context, projectID uint64, path string, recursive bool, page, pageSize int) ([]*entity.File, int64, error)
	CreateFolder(ctx context.Context, projectID, userID uint64, path, folderName string) (*entity.File, error)
	DeleteFile(ctx context.Context, fileID, userID uint64) error
	RestoreFile(ctx context.Context, fileID, userID uint64) error
	GetFileInfo(ctx context.Context, fileID uint64) (*entity.File, error)

	// 版本管理
	GetFileVersions(ctx context.Context, fileID uint64) ([]*entity.FileVersion, error)
	GetFileVersion(ctx context.Context, fileID uint64, version int) (*entity.FileVersion, error)

	// 文件分享
	CreateShare(ctx context.Context, fileID, userID uint64, password string, expireHours, downloadLimit int) (*entity.FileShare, error)
	GetShareInfo(ctx context.Context, shareCode string) (*entity.FileShare, error)
	DownloadSharedFile(ctx context.Context, shareCode, password string) (io.ReadCloser, *entity.File, error)
}

// fileService 文件服务实现
type fileService struct {
	fileRepo    repository.FileRepository
	projectRepo repository.ProjectRepository
	minioClient *minio.Client
	authService AuthService
	db          *gorm.DB
}

// NewFileService 创建文件服务实例
func NewFileService(
	fileRepo repository.FileRepository,
	projectRepo repository.ProjectRepository,
	minioClient *minio.Client,
	authService AuthService,
	db *gorm.DB,
) FileService {
	return &fileService{
		fileRepo:    fileRepo,
		projectRepo: projectRepo,
		minioClient: minioClient,
		authService: authService,
		db:          db,
	}
}

// Upload 上传文件
func (s *fileService) Upload(ctx context.Context, projectID, uploaderID uint64, file *multipart.FileHeader, path string) (*entity.File, error) {
	// 1. 获取项目信息，检查项目是否存在
	project, err := s.projectRepo.GetProjectByID(ctx, projectID)
	if err != nil {
		return nil, err
	}
	if project == nil {
		return nil, errors.New("项目不存在")
	}

	// 确保路径以/结尾
	if path != "" && !strings.HasSuffix(path, "/") {
		path = path + "/"
	}

	// 2. 打开文件
	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("打开文件失败: %w", err)
	}
	defer src.Close()

	// 3. 计算文件哈希值
	fileHash, err := s.minioClient.GetFileHash(src)
	if err != nil {
		return nil, fmt.Errorf("计算文件哈希失败: %w", err)
	}

	// 重置文件指针
	_, err = src.Seek(0, io.SeekStart)
	if err != nil {
		return nil, fmt.Errorf("重置文件指针失败: %w", err)
	}

	// 4. 判断是否可以秒传
	existingFile, err := s.fileRepo.GetByHash(ctx, fileHash)
	if err != nil {
		return nil, fmt.Errorf("查询文件哈希失败: %w", err)
	}

	// 5. 构建文件对象
	fileName := filepath.Base(file.Filename)
	extension := filepath.Ext(fileName)
	fullPath := path + fileName

	// 检查文件名是否在当前目录下已存在
	existingFileAtPath, err := s.fileRepo.GetByPath(ctx, projectID, path, fileName)
	if err != nil {
		return nil, fmt.Errorf("检查文件路径失败: %w", err)
	}

	// 如果同名文件已存在，则创建新版本
	if existingFileAtPath != nil {
		// 创建新版本
		newVersion := &entity.FileVersion{
			FileID:     existingFileAtPath.ID,
			Version:    existingFileAtPath.CurrentVersion + 1,
			FileHash:   fileHash,
			FileSize:   file.Size,
			UploaderID: uploaderID,
			Comment:    "更新文件",
		}

		// 开始事务
		tx := s.db.Begin()
		if tx.Error != nil {
			return nil, tx.Error
		}

		// 事务中添加版本记录
		err = s.fileRepo.CreateVersion(ctx, newVersion)
		if err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("创建版本记录失败: %w", err)
		}

		// 更新文件记录
		existingFileAtPath.FileHash = fileHash
		existingFileAtPath.FileSize = file.Size
		existingFileAtPath.CurrentVersion = newVersion.Version
		existingFileAtPath.UpdatedAt = time.Now()

		err = s.fileRepo.Update(ctx, existingFileAtPath)
		if err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("更新文件记录失败: %w", err)
		}

		// 如果不是秒传，则需要上传文件
		if existingFile == nil {
			// 在MinIO中创建文件
			objectName := minio.GetObjectName(projectID, path, fileName)
			_, err = s.minioClient.UploadFile(ctx, project.Group.GroupKey, objectName, src, file.Size, file.Header.Get("Content-Type"))
			if err != nil {
				tx.Rollback()
				return nil, fmt.Errorf("上传文件失败: %w", err)
			}
		}

		// 提交事务
		if err := tx.Commit().Error; err != nil {
			return nil, fmt.Errorf("提交事务失败: %w", err)
		}

		return existingFileAtPath, nil
	}

	// 6. 创建新文件记录
	newFile := &entity.File{
		ProjectID:      projectID,
		FileName:       fileName,
		FilePath:       path,
		FullPath:       fullPath,
		FileHash:       fileHash,
		FileSize:       file.Size,
		MimeType:       file.Header.Get("Content-Type"),
		Extension:      extension,
		IsFolder:       false,
		UploaderID:     uploaderID,
		CurrentVersion: 1,
	}

	// 开始事务
	tx := s.db.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}

	// 创建文件记录
	err = s.fileRepo.Create(ctx, newFile)
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("创建文件记录失败: %w", err)
	}

	// 创建版本记录
	version := &entity.FileVersion{
		FileID:     newFile.ID,
		Version:    1,
		FileHash:   fileHash,
		FileSize:   file.Size,
		UploaderID: uploaderID,
		Comment:    "初始版本",
	}

	err = s.fileRepo.CreateVersion(ctx, version)
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("创建版本记录失败: %w", err)
	}

	// 如果不是秒传，则需要上传文件
	if existingFile == nil {
		// 在MinIO中创建文件
		objectName := minio.GetObjectName(projectID, path, fileName)
		_, err = s.minioClient.UploadFile(ctx, project.Group.GroupKey, objectName, src, file.Size, file.Header.Get("Content-Type"))
		if err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("上传文件失败: %w", err)
		}
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("提交事务失败: %w", err)
	}

	return newFile, nil
}

// Download 下载文件
func (s *fileService) Download(ctx context.Context, fileID, userID uint64) (io.ReadCloser, *entity.File, error) {
	// 1. 获取文件信息
	file, err := s.fileRepo.GetByID(ctx, fileID)
	if err != nil {
		return nil, nil, err
	}
	if file == nil {
		return nil, nil, errors.New("文件不存在")
	}

	// 2. 检查文件是否已被删除
	if file.IsDeleted {
		return nil, nil, errors.New("文件已被删除")
	}

	// 3. 获取项目信息
	project, err := s.projectRepo.GetProjectByID(ctx, file.ProjectID)
	if err != nil {
		return nil, nil, err
	}
	if project == nil {
		return nil, nil, errors.New("项目不存在")
	}

	// 4. 从MinIO下载文件
	objectName := minio.GetObjectName(file.ProjectID, file.FilePath, file.FileName)
	fileReader, _, err := s.minioClient.DownloadFile(ctx, project.Group.GroupKey, objectName)
	if err != nil {
		return nil, nil, fmt.Errorf("下载文件失败: %w", err)
	}

	return fileReader, file, nil
}

// ListFiles 获取文件列表
func (s *fileService) ListFiles(ctx context.Context, projectID uint64, path string, recursive bool, page, pageSize int) ([]*entity.File, int64, error) {
	// 检查项目是否存在
	project, err := s.projectRepo.GetProjectByID(ctx, projectID)
	if err != nil {
		return nil, 0, err
	}
	if project == nil {
		return nil, 0, errors.New("项目不存在")
	}

	// 获取文件列表
	return s.fileRepo.List(ctx, projectID, path, recursive, false, page, pageSize)
}

// CreateFolder 创建文件夹
func (s *fileService) CreateFolder(ctx context.Context, projectID, userID uint64, path, folderName string) (*entity.File, error) {
	// 1. 获取项目信息
	project, err := s.projectRepo.GetProjectByID(ctx, projectID)
	if err != nil {
		return nil, err
	}
	if project == nil {
		return nil, errors.New("项目不存在")
	}

	// 确保路径以/结尾
	if path != "" && !strings.HasSuffix(path, "/") {
		path = path + "/"
	}

	// 确保文件夹名称不含/
	folderName = strings.TrimSuffix(folderName, "/")
	if strings.Contains(folderName, "/") {
		return nil, errors.New("文件夹名称不能包含'/'")
	}

	// 检查文件夹是否已存在
	fullPath := path + folderName + "/"
	existingFolder, err := s.fileRepo.GetByPath(ctx, projectID, path, folderName)
	if err != nil {
		return nil, fmt.Errorf("检查文件夹是否存在失败: %w", err)
	}
	if existingFolder != nil {
		return nil, errors.New("同名文件夹已存在")
	}

	// 2. 创建文件夹记录
	folder := &entity.File{
		ProjectID:      projectID,
		FileName:       folderName,
		FilePath:       path,
		FullPath:       fullPath,
		FileHash:       "",
		FileSize:       0,
		MimeType:       "application/directory",
		Extension:      "",
		IsFolder:       true,
		UploaderID:     userID,
		CurrentVersion: 1,
	}

	// 3. 在MinIO中创建文件夹
	objectName := minio.GetObjectName(projectID, path, folderName) + "/"
	err = s.minioClient.CreateFolder(ctx, project.Group.GroupKey, objectName)
	if err != nil {
		return nil, fmt.Errorf("创建文件夹失败: %w", err)
	}

	// 4. 保存到数据库
	err = s.fileRepo.Create(ctx, folder)
	if err != nil {
		return nil, fmt.Errorf("保存文件夹记录失败: %w", err)
	}

	return folder, nil
}

// DeleteFile 删除文件(软删除)
func (s *fileService) DeleteFile(ctx context.Context, fileID, userID uint64) error {
	// 1. 获取文件信息
	file, err := s.fileRepo.GetByID(ctx, fileID)
	if err != nil {
		return err
	}
	if file == nil {
		return errors.New("文件不存在")
	}

	// 2. 检查文件是否已被删除
	if file.IsDeleted {
		return errors.New("文件已被删除")
	}

	// 3. 软删除文件
	file.IsDeleted = true
	file.DeletedAt = new(time.Time)
	*file.DeletedAt = time.Now()
	file.DeletedBy = &userID

	err = s.fileRepo.Update(ctx, file)
	if err != nil {
		return fmt.Errorf("删除文件失败: %w", err)
	}

	return nil
}

// RestoreFile 恢复文件
func (s *fileService) RestoreFile(ctx context.Context, fileID, userID uint64) error {
	// 1. 获取文件信息
	file, err := s.fileRepo.GetByID(ctx, fileID)
	if err != nil {
		return err
	}
	if file == nil {
		return errors.New("文件不存在")
	}

	// 2. 检查文件是否已被删除
	if !file.IsDeleted {
		return errors.New("文件未被删除")
	}

	// 3. 恢复文件
	file.IsDeleted = false
	file.DeletedAt = nil
	file.DeletedBy = nil

	err = s.fileRepo.Update(ctx, file)
	if err != nil {
		return fmt.Errorf("恢复文件失败: %w", err)
	}

	return nil
}

// GetFileInfo 获取文件信息
func (s *fileService) GetFileInfo(ctx context.Context, fileID uint64) (*entity.File, error) {
	return s.fileRepo.GetByID(ctx, fileID)
}

// GetFileVersions 获取文件版本列表
func (s *fileService) GetFileVersions(ctx context.Context, fileID uint64) ([]*entity.FileVersion, error) {
	// 1. 获取文件信息
	file, err := s.fileRepo.GetByID(ctx, fileID)
	if err != nil {
		return nil, err
	}
	if file == nil {
		return nil, errors.New("文件不存在")
	}

	// 2. 获取文件版本列表
	return s.fileRepo.GetVersions(ctx, fileID)
}

// GetFileVersion 获取文件特定版本
func (s *fileService) GetFileVersion(ctx context.Context, fileID uint64, version int) (*entity.FileVersion, error) {
	// 1. 获取文件信息
	file, err := s.fileRepo.GetByID(ctx, fileID)
	if err != nil {
		return nil, err
	}
	if file == nil {
		return nil, errors.New("文件不存在")
	}

	// 2. 获取指定版本
	return s.fileRepo.GetVersionByID(ctx, fileID, version)
}

// generateShareCode 生成分享码
func generateShareCode() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const codeLength = 8

	b := make([]byte, codeLength)
	_, err := rand.Read(b)
	if err != nil {
		return uuid.New().String()[:8] // 如果随机数生成失败，使用UUID
	}

	for i := range b {
		b[i] = charset[int(b[i])%len(charset)]
	}

	return string(b)
}

// CreateShare 创建文件分享
func (s *fileService) CreateShare(ctx context.Context, fileID, userID uint64, password string, expireHours, downloadLimit int) (*entity.FileShare, error) {
	// 1. 获取文件信息
	file, err := s.fileRepo.GetByID(ctx, fileID)
	if err != nil {
		return nil, err
	}
	if file == nil {
		return nil, errors.New("文件不存在")
	}

	// 2. 检查文件是否已被删除
	if file.IsDeleted {
		return nil, errors.New("文件已被删除")
	}

	// 3. 创建分享记录
	share := &entity.FileShare{
		FileID:        fileID,
		UserID:        userID,
		ShareCode:     generateShareCode(),
		Password:      password,
		DownloadLimit: downloadLimit,
		DownloadCount: 0,
		CreatedAt:     time.Now(),
	}

	// 设置过期时间
	if expireHours > 0 {
		expireTime := time.Now().Add(time.Duration(expireHours) * time.Hour)
		share.ExpireAt = &expireTime
	}

	// 保存分享记录
	err = s.fileRepo.CreateShare(ctx, share)
	if err != nil {
		return nil, fmt.Errorf("创建分享记录失败: %w", err)
	}

	return share, nil
}

// GetShareInfo 获取分享信息
func (s *fileService) GetShareInfo(ctx context.Context, shareCode string) (*entity.FileShare, error) {
	// 获取分享记录
	share, err := s.fileRepo.GetShareByCode(ctx, shareCode)
	if err != nil {
		return nil, err
	}
	if share == nil {
		return nil, errors.New("分享不存在或已过期")
	}

	// 检查是否过期
	if share.ExpireAt != nil && share.ExpireAt.Before(time.Now()) {
		return nil, errors.New("分享已过期")
	}

	// 检查下载次数是否达到限制
	if share.DownloadLimit > 0 && share.DownloadCount >= share.DownloadLimit {
		return nil, errors.New("分享已达到下载次数限制")
	}

	return share, nil
}

// DownloadSharedFile 下载分享文件
func (s *fileService) DownloadSharedFile(ctx context.Context, shareCode, password string) (io.ReadCloser, *entity.File, error) {
	// 1. 获取分享信息
	share, err := s.GetShareInfo(ctx, shareCode)
	if err != nil {
		return nil, nil, err
	}

	// 2. 检查密码
	if share.Password != "" && share.Password != password {
		return nil, nil, errors.New("密码错误")
	}

	// 3. 获取文件信息
	file, err := s.fileRepo.GetByID(ctx, share.FileID)
	if err != nil {
		return nil, nil, err
	}
	if file == nil {
		return nil, nil, errors.New("文件不存在")
	}

	// 4. 检查文件是否已被删除
	if file.IsDeleted {
		return nil, nil, errors.New("文件已被删除")
	}

	// 5. 获取项目信息
	project, err := s.projectRepo.GetProjectByID(ctx, file.ProjectID)
	if err != nil {
		return nil, nil, err
	}
	if project == nil {
		return nil, nil, errors.New("项目不存在")
	}

	// 6. 更新下载次数
	err = s.fileRepo.UpdateShareDownloadCount(ctx, share.ID)
	if err != nil {
		return nil, nil, fmt.Errorf("更新下载次数失败: %w", err)
	}

	// 7. 从MinIO下载文件
	objectName := minio.GetObjectName(file.ProjectID, file.FilePath, file.FileName)
	fileReader, _, err := s.minioClient.DownloadFile(ctx, project.Group.GroupKey, objectName)
	if err != nil {
		return nil, nil, fmt.Errorf("下载文件失败: %w", err)
	}

	return fileReader, file, nil
}
