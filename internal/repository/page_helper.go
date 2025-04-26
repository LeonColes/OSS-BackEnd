package repository

import (
	"oss-backend/internal/model/dto"

	"gorm.io/gorm"
)

// ApplyPagination 应用分页参数到查询
func ApplyPagination(query *gorm.DB, pageQuery dto.PageQuery) *gorm.DB {
	// 处理页码和页大小
	page := pageQuery.Page
	if page <= 0 {
		page = 1
	}

	size := pageQuery.Size
	if size <= 0 {
		size = 10 // 默认每页10条
	}

	// 计算偏移量
	offset := (page - 1) * size

	// 应用排序
	if pageQuery.SortBy != "" {
		order := pageQuery.SortBy
		if pageQuery.SortOrder == "desc" {
			order += " DESC"
		} else {
			order += " ASC"
		}
		query = query.Order(order)
	} else {
		// 默认按ID降序排序
		query = query.Order("id DESC")
	}

	// 应用分页
	return query.Offset(offset).Limit(size)
}

// GetTotalCount 获取查询的总记录数
func GetTotalCount(query *gorm.DB) (int64, error) {
	var count int64
	// 创建计数查询的副本
	countQuery := query.Session(&gorm.Session{})
	result := countQuery.Count(&count)
	return count, result.Error
}

// ExecutePageQuery 执行分页查询并返回结果
func ExecutePageQuery(query *gorm.DB, pageQuery dto.PageQuery, result interface{}) (int64, error) {
	// 获取总记录数
	total, err := GetTotalCount(query)
	if err != nil {
		return 0, err
	}

	// 应用分页并执行查询
	err = ApplyPagination(query, pageQuery).Find(result).Error
	if err != nil {
		return 0, err
	}

	return total, nil
}
