package repository

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"oss-backend/internal/model/entity"
)

// StorageStatRepository 存储统计仓库接口
type StorageStatRepository interface {
	// 基本CRUD操作
	Create(ctx context.Context, stat *entity.StorageStat) error
	GetByID(ctx context.Context, id string) (*entity.StorageStat, error)
	Update(ctx context.Context, stat *entity.StorageStat) error

	// 特定查询方法
	GetLatestByProject(ctx context.Context, projectID string) (*entity.StorageStat, error)
	GetByDateRange(ctx context.Context, projectID string, startDate, endDate time.Time) ([]*entity.StorageStat, error)
	GetProjectStatsByDate(ctx context.Context, date time.Time) ([]*entity.StorageStat, error)
	GetGroupStatsByDate(ctx context.Context, groupID string, date time.Time) ([]*entity.StorageStat, error)

	// 统计查询方法
	GetProjectTotalStats(ctx context.Context, projectID string) (fileCount int64, totalSize int64, err error)
}

// storageStatRepository 存储统计仓库实现
type storageStatRepository struct {
	db interface {
		WithContext(ctx context.Context) *gorm.DB
	}
}

// NewStorageStatRepository 创建存储统计仓库实例
func NewStorageStatRepository(db *gorm.DB) StorageStatRepository {
	return &storageStatRepository{
		db: db,
	}
}

// Create 创建存储统计记录
func (r *storageStatRepository) Create(ctx context.Context, stat *entity.StorageStat) error {
	return r.db.WithContext(ctx).Create(stat).Error
}

// GetByID 根据ID获取存储统计
func (r *storageStatRepository) GetByID(ctx context.Context, id string) (*entity.StorageStat, error) {
	var stat entity.StorageStat
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&stat).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &stat, nil
}

// Update 更新存储统计记录
func (r *storageStatRepository) Update(ctx context.Context, stat *entity.StorageStat) error {
	return r.db.WithContext(ctx).Save(stat).Error
}

// GetLatestByProject 获取项目最新的存储统计
func (r *storageStatRepository) GetLatestByProject(ctx context.Context, projectID string) (*entity.StorageStat, error) {
	var stat entity.StorageStat
	err := r.db.WithContext(ctx).
		Where("project_id = ?", projectID).
		Order("stat_date DESC").
		First(&stat).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &stat, nil
}

// GetByDateRange 获取指定日期范围的存储统计
func (r *storageStatRepository) GetByDateRange(ctx context.Context, projectID string, startDate, endDate time.Time) ([]*entity.StorageStat, error) {
	var stats []*entity.StorageStat
	err := r.db.WithContext(ctx).
		Where("project_id = ? AND stat_date BETWEEN ? AND ?", projectID, startDate, endDate).
		Order("stat_date ASC").
		Find(&stats).Error
	return stats, err
}

// GetProjectStatsByDate 获取指定日期的所有项目统计
func (r *storageStatRepository) GetProjectStatsByDate(ctx context.Context, date time.Time) ([]*entity.StorageStat, error) {
	var stats []*entity.StorageStat
	err := r.db.WithContext(ctx).
		Where("stat_date = ?", date).
		Find(&stats).Error
	return stats, err
}

// GetGroupStatsByDate 获取指定群组和日期的项目统计
func (r *storageStatRepository) GetGroupStatsByDate(ctx context.Context, groupID string, date time.Time) ([]*entity.StorageStat, error) {
	var stats []*entity.StorageStat
	err := r.db.WithContext(ctx).
		Where("group_id = ? AND stat_date = ?", groupID, date).
		Find(&stats).Error
	return stats, err
}

// GetProjectTotalStats 获取项目当前的总文件数和大小
func (r *storageStatRepository) GetProjectTotalStats(ctx context.Context, projectID string) (fileCount int64, totalSize int64, err error) {
	// 计算文件数
	err = r.db.WithContext(ctx).Model(&entity.File{}).
		Where("project_id = ? AND is_deleted = ? AND is_folder = ?", projectID, false, false).
		Count(&fileCount).Error
	if err != nil {
		return 0, 0, err
	}

	// 计算总大小
	type Result struct {
		TotalSize int64
	}
	var result Result
	err = r.db.WithContext(ctx).Model(&entity.File{}).
		Select("COALESCE(SUM(file_size), 0) as total_size").
		Where("project_id = ? AND is_deleted = ? AND is_folder = ?", projectID, false, false).
		Scan(&result).Error
	if err != nil {
		return fileCount, 0, err
	}

	return fileCount, result.TotalSize, nil
}
