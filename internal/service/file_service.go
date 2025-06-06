package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"oss-backend/internal/model/entity"
	"oss-backend/internal/repository"
	"oss-backend/internal/utils"
	"oss-backend/pkg/minio"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"log"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// FileService 文件服务接口
type FileService interface {
	// 文件操作
	Upload(ctx context.Context, projectID, uploaderID string, file *multipart.FileHeader, path string) (*entity.File, error)
	Download(ctx context.Context, fileID string, userID string) (io.ReadCloser, *entity.File, error)
	ListFiles(ctx context.Context, projectID string, path string, recursive bool, page, pageSize int) ([]*entity.File, int64, error)
	CreateFolder(ctx context.Context, projectID, userID string, path, folderName string) (*entity.File, error)
	DeleteFile(ctx context.Context, fileID, userID string) error
	RestoreFile(ctx context.Context, fileID, userID string) error
	GetFileInfo(ctx context.Context, fileID string) (*entity.File, error)

	// 版本管理
	GetFileVersions(ctx context.Context, fileID string) ([]*entity.FileVersion, error)
	GetFileVersion(ctx context.Context, fileID string, version int) (*entity.FileVersion, error)

	// 文件分享
	CreateShare(ctx context.Context, fileID, userID string, password string, expireHours, downloadLimit int) (*entity.FileShare, error)
	GetShareInfo(ctx context.Context, shareCode string) (*entity.FileShare, error)
	DownloadSharedFile(ctx context.Context, shareCode, password string) (io.ReadCloser, *entity.File, error)

	// 公共下载
	GetPublicDownloadURL(ctx context.Context, fileID string) (string, error)

	// 文件权限
	CheckFilePermission(ctx context.Context, fileID, userID string, requiredAction string) (bool, error)

	// 存储统计
	UpdateStorageStats(ctx context.Context, projectID string, fileSize int64, isAdd bool) error
	RecalculateProjectStats(ctx context.Context, projectID string) error
	VerifyAllProjectsStats(ctx context.Context) error
}

// fileService 文件服务实现
type fileService struct {
	fileRepo    repository.FileRepository
	projectRepo repository.ProjectRepository
	statRepo    repository.StorageStatRepository
	minioClient *minio.Client
	authService AuthService
	db          *gorm.DB
}

// NewFileService 创建文件服务实例
func NewFileService(
	fileRepo repository.FileRepository,
	projectRepo repository.ProjectRepository,
	statRepo repository.StorageStatRepository,
	minioClient *minio.Client,
	authService AuthService,
	db *gorm.DB,
) FileService {
	return &fileService{
		fileRepo:    fileRepo,
		projectRepo: projectRepo,
		statRepo:    statRepo,
		minioClient: minioClient,
		authService: authService,
		db:          db,
	}
}

// Upload 上传文件
func (s *fileService) Upload(ctx context.Context, projectID, uploaderID string, file *multipart.FileHeader, path string) (*entity.File, error) {
	// 1. 获取项目信息，检查项目是否存在
	project, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		return nil, err
	}
	if project == nil {
		return nil, errors.New("项目不存在")
	}

	// 获取群组信息，确认存储桶名称
	if project.Group.GroupKey == "" {
		return nil, errors.New("项目未关联有效群组")
	}
	bucketName := s.sanitizeBucketName(project.Group.GroupKey)

	// 确保存储桶存在
	if err := s.ensureBucketExists(ctx, bucketName); err != nil {
		return nil, fmt.Errorf("存储准备失败: %w", err)
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
	fileHash, err := calculateFileHash(src)
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

		// 计算文件大小差异，用于统计更新
		sizeDiff := file.Size - existingFileAtPath.FileSize

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
			_, err = s.minioClient.UploadFile(ctx, bucketName, objectName, src, file.Size, file.Header.Get("Content-Type"))
			if err != nil {
				tx.Rollback()
				return nil, fmt.Errorf("上传文件失败: %w", err)
			}
		}

		// 提交事务
		if err := tx.Commit().Error; err != nil {
			return nil, fmt.Errorf("提交事务失败: %w", err)
		}

		// 如果文件大小有变化，更新存储统计
		if sizeDiff != 0 {
			// 这里采用异步方式更新统计，避免阻塞主流程
			go func() {
				ctx := context.Background()
				isAdd := sizeDiff > 0
				size := sizeDiff
				if !isAdd {
					size = -sizeDiff
				}
				err := s.UpdateStorageStats(ctx, projectID, size, isAdd)
				if err != nil {
					log.Printf("更新存储统计失败: %v", err)
				}
			}()
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
		_, err = s.minioClient.UploadFile(ctx, bucketName, objectName, src, file.Size, file.Header.Get("Content-Type"))
		if err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("上传文件失败: %w", err)
		}
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("提交事务失败: %w", err)
	}

	// 更新存储统计（异步进行，不阻塞主流程）
	go func() {
		ctx := context.Background()
		err := s.UpdateStorageStats(ctx, projectID, file.Size, true)
		if err != nil {
			log.Printf("更新存储统计失败: %v", err)
		}
	}()

	return newFile, nil
}

// Download 下载文件
func (s *fileService) Download(ctx context.Context, fileID, userID string) (io.ReadCloser, *entity.File, error) {
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
	project, err := s.projectRepo.GetByID(ctx, file.ProjectID)
	if err != nil {
		return nil, nil, fmt.Errorf("获取项目信息失败: %w", err)
	}
	if project == nil {
		return nil, nil, errors.New("项目不存在")
	}

	// 4. 从MinIO下载文件
	objectName := minio.GetObjectName(file.ProjectID, file.FilePath, file.FileName)
	bucketName := s.sanitizeBucketName(project.Group.GroupKey)
	fileReader, _, err := s.minioClient.DownloadFile(ctx, bucketName, objectName)
	if err != nil {
		return nil, nil, fmt.Errorf("下载文件失败: %w", err)
	}

	return fileReader, file, nil
}

// ListFiles 获取文件列表
func (s *fileService) ListFiles(ctx context.Context, projectID string, path string, recursive bool, page, pageSize int) ([]*entity.File, int64, error) {
	// 检查项目是否存在
	project, err := s.projectRepo.GetByID(ctx, projectID)
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
func (s *fileService) CreateFolder(ctx context.Context, projectID, userID string, path, folderName string) (*entity.File, error) {
	// 1. 获取项目信息，检查项目是否存在
	project, err := s.projectRepo.GetByID(ctx, projectID)
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
func (s *fileService) DeleteFile(ctx context.Context, fileID, userID string) error {
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

	// 记录文件大小，用于统计更新
	fileSize := file.FileSize
	projectID := file.ProjectID

	// 3. 软删除文件
	file.IsDeleted = true
	file.DeletedAt = new(time.Time)
	*file.DeletedAt = time.Now()
	file.DeletedBy = &userID

	err = s.fileRepo.Update(ctx, file)
	if err != nil {
		return fmt.Errorf("删除文件失败: %w", err)
	}

	// 异步更新存储统计
	if !file.IsFolder && fileSize > 0 {
		go func() {
			ctx := context.Background()
			err := s.UpdateStorageStats(ctx, projectID, fileSize, false)
			if err != nil {
				log.Printf("更新存储统计失败: %v", err)
			}
		}()
	}

	return nil
}

// RestoreFile 恢复文件
func (s *fileService) RestoreFile(ctx context.Context, fileID, userID string) error {
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

	// 记录文件大小，用于统计更新
	fileSize := file.FileSize
	projectID := file.ProjectID

	// 3. 恢复文件
	file.IsDeleted = false
	file.DeletedAt = nil
	file.DeletedBy = nil

	err = s.fileRepo.Update(ctx, file)
	if err != nil {
		return fmt.Errorf("恢复文件失败: %w", err)
	}

	// 异步更新存储统计
	if !file.IsFolder && fileSize > 0 {
		go func() {
			ctx := context.Background()
			err := s.UpdateStorageStats(ctx, projectID, fileSize, true)
			if err != nil {
				log.Printf("更新存储统计失败: %v", err)
			}
		}()
	}

	return nil
}

// GetFileInfo 获取文件信息
func (s *fileService) GetFileInfo(ctx context.Context, fileID string) (*entity.File, error) {
	return s.fileRepo.GetByID(ctx, fileID)
}

// GetFileVersions 获取文件版本列表
func (s *fileService) GetFileVersions(ctx context.Context, fileID string) ([]*entity.FileVersion, error) {
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
func (s *fileService) GetFileVersion(ctx context.Context, fileID string, version int) (*entity.FileVersion, error) {
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
func (s *fileService) CreateShare(ctx context.Context, fileID, userID string, password string, expireHours, downloadLimit int) (*entity.FileShare, error) {
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
	project, err := s.projectRepo.GetByID(ctx, file.ProjectID)
	if err != nil {
		return nil, nil, err
	}
	if project == nil {
		return nil, nil, errors.New("项目不存在")
	}

	// 6. 从MinIO下载文件
	objectName := minio.GetObjectName(file.ProjectID, file.FilePath, file.FileName)
	bucketName := s.sanitizeBucketName(project.Group.GroupKey)
	fileReader, _, err := s.minioClient.DownloadFile(ctx, bucketName, objectName)
	if err != nil {
		return nil, nil, fmt.Errorf("下载文件失败: %w", err)
	}

	// 7. 更新下载次数
	err = s.fileRepo.UpdateShareDownloadCount(ctx, share.ID)
	if err != nil {
		// 非致命错误，可以忽略
		log.Printf("更新分享下载次数失败: %v", err)
	}

	return fileReader, file, nil
}

// GetPublicDownloadURL 获取公共下载URL
func (s *fileService) GetPublicDownloadURL(ctx context.Context, fileID string) (string, error) {
	// 1. 获取文件信息
	file, err := s.fileRepo.GetByID(ctx, fileID)
	if err != nil {
		return "", err
	}
	if file == nil {
		return "", errors.New("文件不存在")
	}

	// 2. 获取项目信息
	project, err := s.projectRepo.GetByID(ctx, file.ProjectID)
	if err != nil {
		return "", fmt.Errorf("获取项目信息失败: %w", err)
	}
	if project == nil {
		return "", errors.New("项目不存在")
	}

	// 3. 生成公共下载URL
	objectName := minio.GetObjectName(file.ProjectID, file.FilePath, file.FileName)
	bucketName := s.sanitizeBucketName(project.Group.GroupKey)
	return s.minioClient.GetPublicDownloadURL(ctx, bucketName, objectName)
}

func (s *fileService) CheckFilePermission(ctx context.Context, fileID, userID string, requiredAction string) (bool, error) {
	// 1. 获取文件信息
	file, err := s.fileRepo.GetByID(ctx, fileID)
	if err != nil {
		return false, err
	}
	if file == nil {
		return false, errors.New("文件不存在")
	}

	// 2. 获取项目信息
	project, err := s.projectRepo.GetByID(ctx, file.ProjectID)
	if err != nil {
		return false, err
	}
	if project == nil {
		return false, errors.New("项目不存在")
	}

	// 3. 检查用户是否拥有执行所需操作的权限
	projectDomain := fmt.Sprintf("project:%s", file.ProjectID)
	return s.authService.CanUserAccessResource(ctx, userID, "files", requiredAction, projectDomain)
}

// ensureBucketExists 确保存储桶存在
func (s *fileService) ensureBucketExists(ctx context.Context, bucketName string) error {
	// bucketName应该已经通过sanitizeBucketName函数处理过了

	// 检查并创建存储桶
	err := s.minioClient.CreateBucketIfNotExists(ctx, bucketName)
	if err != nil {
		log.Printf("确保存储桶 %s 存在时发生错误: %v", bucketName, err)
		// 检查是否是网络问题
		if strings.Contains(err.Error(), "connection") || strings.Contains(err.Error(), "timeout") {
			return fmt.Errorf("连接MinIO服务器失败，请检查网络或服务器状态: %w", err)
		}
		// 检查是否是权限问题
		if strings.Contains(err.Error(), "AccessDenied") || strings.Contains(err.Error(), "access denied") {
			return fmt.Errorf("MinIO权限被拒绝，请联系管理员创建存储桶 %s 或调整权限: %w", bucketName, err)
		}
		return fmt.Errorf("创建MinIO存储桶失败: %w", err)
	}
	return nil
}

// sanitizeBucketName 规范化桶名称，使其符合S3规范
func (s *fileService) sanitizeBucketName(key string) string {
	// 生成符合S3规范的桶名称：只能包含小写字母、数字和连字符
	// 1. 将所有字符转为小写
	lowerKey := strings.ToLower(key)
	// 2. 替换所有非法字符为连字符
	reg := regexp.MustCompile(`[^a-z0-9\-]`)
	sanitizedKey := reg.ReplaceAllString(lowerKey, "-")
	// 3. 确保不以连字符开头或结尾
	sanitizedKey = strings.Trim(sanitizedKey, "-")
	// 4. 如果长度不足，添加前缀
	if len(sanitizedKey) < 3 {
		sanitizedKey = fmt.Sprintf("grp-%s", sanitizedKey)
	}
	// 5. 如果长度过长，截断
	if len(sanitizedKey) > 60 {
		sanitizedKey = sanitizedKey[:60]
	}

	return fmt.Sprintf("group-%s", sanitizedKey)
}

// calculateFileHash 计算文件哈希
func calculateFileHash(reader io.Reader) (string, error) {
	hash := sha256.New()
	if _, err := io.Copy(hash, reader); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

// UpdateStorageStats 更新存储统计
func (s *fileService) UpdateStorageStats(ctx context.Context, projectID string, fileSize int64, isAdd bool) error {
	today := time.Now().Truncate(24 * time.Hour)

	// 事务操作
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 先获取项目信息
		project, err := s.projectRepo.GetByID(ctx, projectID)
		if err != nil {
			return fmt.Errorf("获取项目信息失败: %w", err)
		}
		if project == nil {
			return errors.New("项目不存在")
		}

		// 查找今日统计记录
		var stat entity.StorageStat
		result := tx.Where("project_id = ? AND stat_date = ?", projectID, today).First(&stat)

		if result.Error != nil {
			if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
				return fmt.Errorf("查询存储统计失败: %w", result.Error)
			}

			// 记录不存在，创建新记录
			// 计算当前文件数和大小
			fileCount, totalSize, err := s.statRepo.GetProjectTotalStats(ctx, projectID)
			if err != nil {
				return fmt.Errorf("计算项目统计失败: %w", err)
			}

			// 创建今天的统计记录
			var increaseValue int64 = 0
			if isAdd {
				increaseValue = fileSize
			}

			stat = entity.StorageStat{
				ID:           utils.GenerateRecordID(),
				GroupID:      project.GroupID,
				ProjectID:    projectID,
				StatDate:     today,
				FileCount:    fileCount,
				TotalSize:    totalSize,
				IncreaseSize: increaseValue, // 如果是添加文件，则增加增量
				CreatedAt:    time.Now(),
			}

			return tx.Create(&stat).Error
		}

		// 更新已有记录
		updates := map[string]interface{}{}

		if isAdd {
			updates["file_count"] = gorm.Expr("file_count + ?", 1)
			updates["total_size"] = gorm.Expr("total_size + ?", fileSize)
			updates["increase_size"] = gorm.Expr("increase_size + ?", fileSize)
		} else {
			updates["file_count"] = gorm.Expr("file_count - ?", 1)
			updates["total_size"] = gorm.Expr("total_size - ?", fileSize)
			// 不减少 increase_size，因为它表示的是一段时间内的增量
		}

		return tx.Model(&entity.StorageStat{}).
			Where("id = ?", stat.ID).
			Updates(updates).Error
	})
}

// RecalculateProjectStats 重新计算项目统计
func (s *fileService) RecalculateProjectStats(ctx context.Context, projectID string) error {
	today := time.Now().Truncate(24 * time.Hour)

	// 事务操作
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 获取项目信息
		project, err := s.projectRepo.GetByID(ctx, projectID)
		if err != nil {
			return fmt.Errorf("获取项目信息失败: %w", err)
		}
		if project == nil {
			return errors.New("项目不存在")
		}

		// 计算当前文件数和大小
		fileCount, totalSize, err := s.statRepo.GetProjectTotalStats(ctx, projectID)
		if err != nil {
			return fmt.Errorf("计算项目统计失败: %w", err)
		}

		// 查找今日统计记录
		var stat entity.StorageStat
		result := tx.Where("project_id = ? AND stat_date = ?", projectID, today).First(&stat)

		// 计算增量值需要昨天的数据
		yesterday := today.AddDate(0, 0, -1)
		var yesterdayStat entity.StorageStat
		tx.Where("project_id = ? AND stat_date = ?", projectID, yesterday).First(&yesterdayStat)

		var increaseSize int64 = 0
		if yesterdayStat.ID != "" {
			// 有昨天的数据，计算增量
			increaseSize = totalSize - yesterdayStat.TotalSize
			if increaseSize < 0 {
				increaseSize = 0 // 防止负值
			}
		} else {
			// 没有昨天的数据，增量就是今天的全部大小
			increaseSize = totalSize
		}

		if result.Error != nil {
			if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
				return fmt.Errorf("查询存储统计失败: %w", result.Error)
			}

			// 记录不存在，创建新记录
			stat = entity.StorageStat{
				ID:           utils.GenerateRecordID(),
				GroupID:      project.GroupID,
				ProjectID:    projectID,
				StatDate:     today,
				FileCount:    fileCount,
				TotalSize:    totalSize,
				IncreaseSize: increaseSize,
				CreatedAt:    time.Now(),
			}

			return tx.Create(&stat).Error
		}

		// 更新已有记录
		stat.FileCount = fileCount
		stat.TotalSize = totalSize
		stat.IncreaseSize = increaseSize

		return tx.Save(&stat).Error
	})
}

// VerifyAllProjectsStats 验证所有项目统计
func (s *fileService) VerifyAllProjectsStats(ctx context.Context) error {
	// 获取所有项目
	projects, err := s.projectRepo.GetAll(ctx)
	if err != nil {
		return fmt.Errorf("获取项目列表失败: %w", err)
	}

	// 逐个重新计算项目统计
	for _, project := range projects {
		err := s.RecalculateProjectStats(ctx, project.ID)
		if err != nil {
			log.Printf("重新计算项目 %s 统计失败: %v", project.ID, err)
			// 继续处理其他项目，不中断
			continue
		}
	}

	return nil
}
