package dto

import "time"

// ===== 请求结构 =====

// GroupCreateRequest 创建群组请求
type GroupCreateRequest struct {
	Name        string `json:"name" binding:"required,max=64"`        // 群组名称
	Description string `json:"description" binding:"max=500"`         // 群组描述
	GroupKey    string `json:"group_key" binding:"required,alphanum"` // 群组标识(仅允许字母和数字)
}

// GroupUpdateRequest 更新群组请求
type GroupUpdateRequest struct {
	ID          string `json:"id" binding:"required"`
	Name        string `json:"name" binding:"required,max=64"`
	Description string `json:"description" binding:"max=500"`
	Status      *int   `json:"status,omitempty"`
}

// GroupListRequest 群组列表请求
type GroupListRequest struct {
	Name      string `form:"name"`            // 群组名称(模糊查询)
	Status    int    `form:"status"`          // 状态:1-正常,2-禁用,3-锁定
	Page      int    `form:"page,default=1"`  // 页码
	Size      int    `form:"size,default=10"` // 每页数量
	CreatorID string `form:"creator_id"`      // 创建者ID，用于筛选特定创建者的群组
	PageSize  int    `form:"page_size"`       // 页面大小别名，与Size等效
	SortBy    string `form:"sort_by"`         // 排序字段
	SortOrder string `form:"sort_order"`      // 排序方式（asc/desc）
}

// GroupJoinRequest 加入群组请求
type GroupJoinRequest struct {
	InviteCode string `json:"invite_code" binding:"required"` // 邀请码
}

// GroupMemberUpdateRequest 更新群组成员角色请求
type GroupMemberUpdateRequest struct {
	UserID string `json:"user_id" binding:"required"`                 // 用户ID
	Role   string `json:"role" binding:"required,oneof=admin member"` // 角色
}

// GroupInviteRequest 生成邀请码请求
type GroupInviteRequest struct {
	GroupID    string `json:"group_id" binding:"required"` // 群组ID
	ExpireDays int    `json:"expire_days,omitempty"`       // 过期天数,0表示永不过期
}

// ===== 响应结构 =====

// GroupResponse 群组响应
type GroupResponse struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	GroupKey     string    `json:"group_key"`
	InviteCode   string    `json:"invite_code,omitempty"` // 仅群组管理员可见
	StorageQuota int64     `json:"storage_quota"`         // 存储配额,0表示无限制
	StorageUsed  int64     `json:"storage_used"`          // 已使用存储量
	MemberCount  int       `json:"member_count"`          // 成员数量
	ProjectCount int       `json:"project_count"`         // 项目数量
	Status       int       `json:"status"`                // 状态:1-正常,2-禁用,3-锁定
	CreatorID    string    `json:"creator_id"`            // 创建者ID
	CreatorName  string    `json:"creator_name"`          // 创建者名称
	CreatedAt    time.Time `json:"created_at"`            // 创建时间
	UserRole     string    `json:"user_role,omitempty"`   // 当前用户在群组中的角色
}

// GroupMemberResponse 群组成员响应
type GroupMemberResponse struct {
	ID           string     `json:"id"`                       // 成员ID
	UserID       string     `json:"user_id"`                  // 用户ID
	UserName     string     `json:"user_name"`                // 用户名称
	Email        string     `json:"email"`                    // 邮箱
	Avatar       string     `json:"avatar"`                   // 头像
	Role         string     `json:"role"`                     // 角色
	JoinedAt     time.Time  `json:"joined_at"`                // 加入时间
	LastActiveAt *time.Time `json:"last_active_at,omitempty"` // 最后活跃时间
}

// GroupInviteResponse 群组邀请响应
type GroupInviteResponse struct {
	GroupID    string     `json:"group_id"`    // 群组ID
	GroupName  string     `json:"group_name"`  // 群组名称
	InviteCode string     `json:"invite_code"` // 邀请码
	ExpireAt   *time.Time `json:"expire_at"`   // 过期时间
}

// GroupListResponse 群组列表响应
type GroupListResponse struct {
	Total int64           `json:"total"` // 总数
	Items []GroupResponse `json:"items"` // 群组列表
}

// GroupMemberListResponse 群组成员列表响应
type GroupMemberListResponse struct {
	Total int64                 `json:"total"` // 总数
	Items []GroupMemberResponse `json:"items"` // 成员列表
}
