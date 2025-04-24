package dto

// RoleCreateRequest 创建角色请求
type RoleCreateRequest struct {
	Name        string `json:"name" binding:"required" example:"管理员"`   // 角色名称
	Description string `json:"description" example:"系统管理员角色"`           // 角色描述
	Code        string `json:"code" binding:"required" example:"ADMIN"` // 角色编码
}

// RoleUpdateRequest 更新角色请求
type RoleUpdateRequest struct {
	ID          uint   `json:"id" binding:"required" example:"1"`       // 角色ID
	Name        string `json:"name" binding:"required" example:"管理员"`   // 角色名称
	Description string `json:"description" example:"系统管理员角色"`           // 角色描述
	Code        string `json:"code" binding:"required" example:"ADMIN"` // 角色编码
	Status      int    `json:"status" example:"1"`                      // 状态：1-启用，0-禁用
}

// RoleResponse 角色响应
type RoleResponse struct {
	ID          uint   `json:"id" example:"1"`                           // 角色ID
	Name        string `json:"name" example:"管理员"`                       // 角色名称
	Description string `json:"description" example:"系统管理员角色"`            // 角色描述
	Code        string `json:"code" example:"ADMIN"`                     // 角色编码
	Status      int    `json:"status" example:"1"`                       // 状态：1-启用，0-禁用
	IsSystem    bool   `json:"is_system" example:"true"`                 // 是否为系统角色
	CreatedAt   string `json:"created_at" example:"2023-01-01 12:00:00"` // 创建时间
}

// RoleListRequest 角色列表请求
type RoleListRequest struct {
	Name   string `form:"name" example:"管理员"` // 角色名称，模糊查询
	Status int    `form:"status" example:"1"` // 状态：1-启用，0-禁用
	Page   int    `form:"page" example:"1"`   // 页码
	Size   int    `form:"size" example:"10"`  // 每页数量
}

// RoleListResponse 角色列表响应
type RoleListResponse struct {
	Total int64          `json:"total" example:"100"` // 总数
	List  []RoleResponse `json:"list"`                // 角色列表
}
