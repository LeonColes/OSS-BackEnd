package repository

import (
	"context"
	"errors"
	"time"

	"oss-backend/internal/model/entity"

	"gorm.io/gorm"
)

// 定义错误
var (
	// 文件存储库特定错误
	ErrFileNotFound        = errors.New("文件不存在")
	ErrFileVersionNotFound = errors.New("文件版本不存在")
	ErrShareNotFound       = errors.New("分享不存在")
	ErrShareExpired        = errors.New("分享已过期")
	ErrFileAlreadyExists   = errors.New("文件已存在")
)

// FileRepository 文件存储库接口
type FileRepository interface {
	// 保存文件信息
	SaveFile(ctx context.Context, file *entity.File) error
	// 根据ID获取文件
	GetFileByID(ctx context.Context, fileID uint64) (*entity.File, error)
	// 获取文件夹下的文件列表
	ListFilesByParent(ctx context.Context, projectID, parentID uint64, userID uint64) ([]*entity.File, error)
	// 更新文件信息
	UpdateFile(ctx context.Context, file *entity.File) error
	// 删除文件（逻辑删除）
	DeleteFile(ctx context.Context, fileID uint64, userID uint64) error
	// 永久删除文件
	PermanentDeleteFile(ctx context.Context, fileID uint64) error
	// 创建文件版本
	CreateFileVersion(ctx context.Context, version *entity.FileVersion) error
	// 获取文件的版本列表
	GetFileVersions(ctx context.Context, fileID uint64) ([]*entity.FileVersion, error)
	// 获取特定文件版本
	GetFileVersion(ctx context.Context, fileID uint64, version int) (*entity.FileVersion, error)
	// 创建分享
	CreateFileShare(ctx context.Context, share *entity.FileShare) error
	// 获取分享信息
	GetFileShareByCode(ctx context.Context, code string) (*entity.FileShare, error)
	// 更新分享信息
	UpdateFileShare(ctx context.Context, share *entity.FileShare) error
	// 删除分享
	DeleteFileShare(ctx context.Context, shareID uint64) error
}

// fileRepository 文件存储库实现
type fileRepository struct {
	db *gorm.DB
}

// NewFileRepository 创建文件存储库
func NewFileRepository(db *gorm.DB) FileRepository {
	return &fileRepository{db: db}
}

// SaveFile 保存文件信息
func (r *fileRepository) SaveFile(ctx context.Context, file *entity.File) error {
	return r.db.Create(file).Error
}

// GetFileByID 根据ID获取文件
func (r *fileRepository) GetFileByID(ctx context.Context, fileID uint64) (*entity.File, error) {
	var file entity.File
	err := r.db.Where("id = ? AND is_deleted = ?", fileID, false).First(&file).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrFileNotFound
		}
		return nil, err
	}
	return &file, nil
}

// ListFilesByParent 获取文件夹下的文件列表
func (r *fileRepository) ListFilesByParent(ctx context.Context, projectID, parentID uint64, userID uint64) ([]*entity.File, error) {
	var files []*entity.File
	// 查询条件：同一项目下、指定父目录、未删除
	err := r.db.Where("project_id = ? AND parent_id = ? AND is_deleted = ?", projectID, parentID, false).Find(&files).Error
	if err != nil {
		return nil, err
	}
	return files, nil
}

// UpdateFile 更新文件信息
func (r *fileRepository) UpdateFile(ctx context.Context, file *entity.File) error {
	return r.db.Save(file).Error
}

// DeleteFile 删除文件（逻辑删除）
func (r *fileRepository) DeleteFile(ctx context.Context, fileID uint64, userID uint64) error {
	// 逻辑删除文件
	now := time.Now()
	return r.db.Model(&entity.File{}).Where("id = ?", fileID).Updates(map[string]interface{}{
		"is_deleted": true,
		"deleted_at": now,
		"deleted_by": userID,
	}).Error
}

// PermanentDeleteFile 永久删除文件
func (r *fileRepository) PermanentDeleteFile(ctx context.Context, fileID uint64) error {
	// 永久删除文件（使用GORM的Delete方法进行硬删除）
	return r.db.Unscoped().Delete(&entity.File{}, fileID).Error
}

// CreateFileVersion 创建文件版本
func (r *fileRepository) CreateFileVersion(ctx context.Context, version *entity.FileVersion) error {
	return r.db.Create(version).Error
}

// GetFileVersions 获取文件的版本列表
func (r *fileRepository) GetFileVersions(ctx context.Context, fileID uint64) ([]*entity.FileVersion, error) {
	var versions []*entity.FileVersion
	err := r.db.Where("file_id = ?", fileID).Order("version desc").Find(&versions).Error
	if err != nil {
		return nil, err
	}
	return versions, nil
}

// GetFileVersion 获取特定文件版本
func (r *fileRepository) GetFileVersion(ctx context.Context, fileID uint64, version int) (*entity.FileVersion, error) {
	var fileVersion entity.FileVersion
	err := r.db.Where("file_id = ? AND version = ?", fileID, version).First(&fileVersion).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrFileVersionNotFound
		}
		return nil, err
	}
	return &fileVersion, nil
}

// CreateFileShare 创建分享
func (r *fileRepository) CreateFileShare(ctx context.Context, share *entity.FileShare) error {
	return r.db.Create(share).Error
}

// GetFileShareByCode 获取分享信息
func (r *fileRepository) GetFileShareByCode(ctx context.Context, code string) (*entity.FileShare, error) {
	var share entity.FileShare
	err := r.db.Where("share_code = ?", code).First(&share).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrShareNotFound
		}
		return nil, err
	}
	return &share, nil
}

// UpdateFileShare 更新分享信息
func (r *fileRepository) UpdateFileShare(ctx context.Context, share *entity.FileShare) error {
	return r.db.Save(share).Error
}

// DeleteFileShare 删除分享
func (r *fileRepository) DeleteFileShare(ctx context.Context, shareID uint64) error {
	return r.db.Delete(&entity.FileShare{}, shareID).Error
}
