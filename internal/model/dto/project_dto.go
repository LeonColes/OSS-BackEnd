package dto

import "time"

// CreateProjectRequest 创建项目请求
type CreateProjectRequest struct {
	Name        string `json:"name" binding:"required,min=2,max=64"`
	Description string `json:"description" binding:"max=500"`
	GroupID     string `json:"group_id" binding:"required"`
}

// UpdateProjectRequest 更新项目请求
type UpdateProjectRequest struct {
	ID          string `json:"id" binding:"required"`
	Name        string `json:"name" binding:"required,min=2,max=64"`
	Description string `json:"description" binding:"max=500"`
	Status      int    `json:"status" binding:"omitempty,oneof=1 2"`
}

// ProjectQuery 项目查询参数
type ProjectQuery struct {
	GroupID string `form:"group_id" binding:"omitempty"`
	Status  int    `form:"status" binding:"omitempty,oneof=1 2 3"`
	Keyword string `form:"keyword" binding:"omitempty,max=50"`
	Page    int    `form:"page" binding:"omitempty,min=1"`
	Size    int    `form:"size" binding:"omitempty,min=5,max=100"`
}

// ProjectResponse 项目响应
type ProjectResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	GroupID     string    `json:"group_id"`
	GroupName   string    `json:"group_name"`
	PathPrefix  string    `json:"path_prefix"`
	CreatorID   string    `json:"creator_id"`
	CreatorName string    `json:"creator_name"`
	Status      int       `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	FileCount   int64     `json:"file_count"`
	TotalSize   int64     `json:"total_size"`
}

// SetPermissionRequest 设置项目权限请求
type SetPermissionRequest struct {
	ProjectID string `json:"project_id" binding:"required"`
	UserID    string `json:"user_id" binding:"required"`
	Role      string `json:"role" binding:"required,oneof=admin editor viewer"`
	ExpireAt  string `json:"expire_at" binding:"omitempty,datetime=2006-01-02 15:04:05"`
}

// RemovePermissionRequest 移除项目权限请求
type RemovePermissionRequest struct {
	ProjectID string `json:"project_id" binding:"required"`
	UserID    string `json:"user_id" binding:"required"`
}

// ProjectUserResponse 项目用户权限响应
type ProjectUserResponse struct {
	UserID      string     `json:"user_id"`
	Username    string     `json:"username"`
	Email       string     `json:"email"`
	Avatar      string     `json:"avatar"`
	Role        string     `json:"role"`
	GrantedBy   string     `json:"granted_by"`
	GranterName string     `json:"granter_name"`
	CreatedAt   time.Time  `json:"created_at"`
	ExpireAt    *time.Time `json:"expire_at"`
}

// ProjectListRequest 项目列表查询请求参数
type ProjectListRequest struct {
	Name      string `json:"name" form:"name"`             // 项目名称（模糊匹配）
	GroupID   string `json:"group_id" form:"group_id"`     // 群组ID
	Status    int    `json:"status" form:"status"`         // 项目状态
	Page      int    `json:"page" form:"page"`             // 页码
	PageSize  int    `json:"page_size" form:"page_size"`   // 每页大小
	SortBy    string `json:"sort_by" form:"sort_by"`       // 排序字段
	SortOrder string `json:"sort_order" form:"sort_order"` // 排序方向（asc/desc）
}
