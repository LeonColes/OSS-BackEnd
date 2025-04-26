package dto

// PageQuery 统一分页查询参数
type PageQuery struct {
	Page      int    `json:"page" form:"page"`             // 页码，从1开始
	Size      int    `json:"size" form:"size"`             // 每页大小
	SortBy    string `json:"sort_by" form:"sort_by"`       // 排序字段
	SortOrder string `json:"sort_order" form:"sort_order"` // 排序方向 (asc/desc)
}

// WithDefaultValues 设置默认值
func (q PageQuery) WithDefaultValues() PageQuery {
	if q.Page <= 0 {
		q.Page = 1
	}
	if q.Size <= 0 {
		q.Size = 10
	}
	return q
}

// PageResult 统一分页响应结构
type PageResult struct {
	List      interface{} `json:"list"`       // 数据列表
	Total     int64       `json:"total"`      // 总记录数
	Page      int         `json:"page"`       // 当前页码
	Size      int         `json:"size"`       // 每页大小
	TotalPage int         `json:"total_page"` // 总页数
}

// NewPageResult 创建分页结果
func NewPageResult(list interface{}, total int64, query PageQuery) PageResult {
	// 计算总页数
	totalPage := 0
	if query.Size > 0 {
		totalPage = int((total + int64(query.Size) - 1) / int64(query.Size))
	}

	return PageResult{
		List:      list,
		Total:     total,
		Page:      query.Page,
		Size:      query.Size,
		TotalPage: totalPage,
	}
}
