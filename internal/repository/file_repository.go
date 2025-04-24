package repository

import (
	"context"
	"errors"
	"oss-backend/internal/model/entity"
	"strings"

	"gorm.io/gorm"
)

// FileRepository 文件仓库接口
type FileRepository interface {
	// 基础CRUD操作
	Create(ctx context.Context, file *entity.File) error
	GetByID(ctx context.Context, id uint64) (*entity.File, error)
	Update(ctx context.Context, file *entity.File) error
	Delete(ctx context.Context, id uint64) error

	// 文件列表操作
	List(ctx context.Context, projectID uint64, path string, recursive bool, includeDeleted bool, page, pageSize int) ([]*entity.File, int64, error)
	ListByIDs(ctx context.Context, ids []uint64) ([]*entity.File, error)

	// 特定查询方法
	GetByHash(ctx context.Context, hash string) (*entity.File, error)
	GetByPath(ctx context.Context, projectID uint64, path string, fileName string) (*entity.File, error)

	// 版本管理
	CreateVersion(ctx context.Context, version *entity.FileVersion) error
	GetVersions(ctx context.Context, fileID uint64) ([]*entity.FileVersion, error)
	GetVersionByID(ctx context.Context, fileID uint64, version int) (*entity.FileVersion, error)

	// 分享管理
	CreateShare(ctx context.Context, share *entity.FileShare) error
	GetShareByCode(ctx context.Context, code string) (*entity.FileShare, error)
	UpdateShareDownloadCount(ctx context.Context, shareID uint64) error
	DeleteShare(ctx context.Context, id uint64) error
}

// fileRepository 文件仓库实现
type fileRepository struct {
	db *gorm.DB
}

// NewFileRepository 创建文件仓库实例
func NewFileRepository(db *gorm.DB) FileRepository {
	return &fileRepository{
		db: db,
	}
}

// Create 创建文件记录
func (r *fileRepository) Create(ctx context.Context, file *entity.File) error {
	return r.db.WithContext(ctx).Create(file).Error
}

// GetByID 根据ID获取文件
func (r *fileRepository) GetByID(ctx context.Context, id uint64) (*entity.File, error) {
	var file entity.File
	err := r.db.WithContext(ctx).First(&file, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &file, nil
}

// Update 更新文件记录
func (r *fileRepository) Update(ctx context.Context, file *entity.File) error {
	return r.db.WithContext(ctx).Save(file).Error
}

// Delete 删除文件（软删除）
func (r *fileRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Model(&entity.File{}).Where("id = ?", id).Update("is_deleted", true).Error
}

// List 获取文件列表
func (r *fileRepository) List(ctx context.Context, projectID uint64, path string, recursive bool, includeDeleted bool, page, pageSize int) ([]*entity.File, int64, error) {
	var files []*entity.File
	var total int64

	// 确保路径以/结尾
	if path != "" && !strings.HasSuffix(path, "/") {
		path = path + "/"
	}

	query := r.db.WithContext(ctx).Model(&entity.File{}).Where("project_id = ?", projectID)

	// 路径筛选
	if path == "" || path == "/" {
		// 根目录，只显示根目录下的文件
		query = query.Where("file_path = ? OR file_path = ?", "", "/")
	} else if recursive {
		// 递归显示子目录
		query = query.Where("file_path LIKE ?", path+"%")
	} else {
		// 只显示当前目录
		query = query.Where("file_path = ?", path)
	}

	// 是否包含已删除文件
	if !includeDeleted {
		query = query.Where("is_deleted = ?", false)
	}

	// 计算总数
	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// 分页查询
	if page > 0 && pageSize > 0 {
		offset := (page - 1) * pageSize
		query = query.Offset(offset).Limit(pageSize)
	}

	// 执行查询
	err = query.Order("is_folder DESC, file_name ASC").Find(&files).Error
	if err != nil {
		return nil, 0, err
	}

	return files, total, nil
}

// ListByIDs 根据ID列表获取文件
func (r *fileRepository) ListByIDs(ctx context.Context, ids []uint64) ([]*entity.File, error) {
	var files []*entity.File
	err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&files).Error
	return files, err
}

// GetByHash 根据文件哈希获取文件
func (r *fileRepository) GetByHash(ctx context.Context, hash string) (*entity.File, error) {
	var file entity.File
	err := r.db.WithContext(ctx).Where("file_hash = ? AND is_deleted = ?", hash, false).First(&file).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &file, nil
}

// GetByPath 根据路径和名称获取文件
func (r *fileRepository) GetByPath(ctx context.Context, projectID uint64, path string, fileName string) (*entity.File, error) {
	var file entity.File

	// 确保路径以/结尾
	if path != "" && !strings.HasSuffix(path, "/") {
		path = path + "/"
	}

	// 构建完整路径
	fullPath := path + fileName

	err := r.db.WithContext(ctx).Where("project_id = ? AND full_path = ? AND is_deleted = ?",
		projectID, fullPath, false).First(&file).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &file, nil
}

// CreateVersion 创建文件版本
func (r *fileRepository) CreateVersion(ctx context.Context, version *entity.FileVersion) error {
	return r.db.WithContext(ctx).Create(version).Error
}

// GetVersions 获取文件所有版本
func (r *fileRepository) GetVersions(ctx context.Context, fileID uint64) ([]*entity.FileVersion, error) {
	var versions []*entity.FileVersion
	err := r.db.WithContext(ctx).Where("file_id = ?", fileID).Order("version DESC").Find(&versions).Error
	return versions, err
}

// GetVersionByID 获取文件特定版本
func (r *fileRepository) GetVersionByID(ctx context.Context, fileID uint64, version int) (*entity.FileVersion, error) {
	var fileVersion entity.FileVersion
	err := r.db.WithContext(ctx).Where("file_id = ? AND version = ?", fileID, version).First(&fileVersion).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &fileVersion, nil
}

// CreateShare 创建文件分享
func (r *fileRepository) CreateShare(ctx context.Context, share *entity.FileShare) error {
	return r.db.WithContext(ctx).Create(share).Error
}

// GetShareByCode 根据分享码获取分享
func (r *fileRepository) GetShareByCode(ctx context.Context, code string) (*entity.FileShare, error) {
	var share entity.FileShare
	err := r.db.WithContext(ctx).Where("share_code = ?", code).First(&share).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	// 预加载文件信息
	err = r.db.WithContext(ctx).Model(&share).Association("File").Find(&share.File)
	if err != nil {
		return nil, err
	}

	return &share, nil
}

// UpdateShareDownloadCount 更新分享下载次数
func (r *fileRepository) UpdateShareDownloadCount(ctx context.Context, shareID uint64) error {
	return r.db.WithContext(ctx).Model(&entity.FileShare{}).Where("id = ?", shareID).
		UpdateColumn("download_count", gorm.Expr("download_count + ?", 1)).Error
}

// DeleteShare 删除分享
func (r *fileRepository) DeleteShare(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&entity.FileShare{}, id).Error
}
