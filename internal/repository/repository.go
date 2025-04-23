package repository

import (
	"gorm.io/gorm"
)

// Repository 定义所有仓库的基础接口
type Repository interface {
	GetDB() *gorm.DB
}

// BaseRepository 提供基本的仓库实现
type BaseRepository struct {
	db *gorm.DB
}

// NewBaseRepository 创建基础仓库实例
func NewBaseRepository(db *gorm.DB) *BaseRepository {
	return &BaseRepository{db: db}
}

// GetDB 获取数据库连接
func (r *BaseRepository) GetDB() *gorm.DB {
	return r.db
} 