package dto

import "time"

// UserRegisterRequest 用户注册请求
type UserRegisterRequest struct {
	Email    string `json:"email" binding:"required,email" example:"user@x.com"`       // 用户邮箱
	Password string `json:"password" binding:"required,min=6,max=20" example:"123456"` // 密码
	Name     string `json:"name" binding:"required" example:"user"`                    // 用户姓名
}

// UserLoginRequest 用户登录请求
type UserLoginRequest struct {
	Email    string `json:"email" binding:"required,email" example:"user@x.com"` // 用户邮箱
	Password string `json:"password" binding:"required" example:"123456"`        // 密码
}

// UserUpdateRequest 用户信息更新请求
type UserUpdateRequest struct {
	Name   string `json:"name" example:"张三"`                               // 用户姓名
	Avatar string `json:"avatar" example:"https://example.com/avatar.jpg"` // 头像URL
}

// UserPasswordUpdateRequest 用户密码更新请求
type UserPasswordUpdateRequest struct {
	OldPassword string `json:"old_password" binding:"required" example:"oldpassword123"`              // 旧密码
	NewPassword string `json:"new_password" binding:"required,min=6,max=20" example:"newpassword123"` // 新密码
}

// UserResponse 用户信息响应
type UserResponse struct {
	ID          string         `json:"id" example:"1"`                                  // 用户ID
	Email       string         `json:"email" example:"user@example.com"`                // 用户邮箱
	Name        string         `json:"name" example:"张三"`                               // 用户姓名
	Avatar      string         `json:"avatar" example:"https://example.com/avatar.jpg"` // 头像URL
	Status      int            `json:"status" example:"1"`                              // 状态：1-正常，2-禁用，3-锁定
	LastLoginAt *time.Time     `json:"last_login_at,omitempty"`                         // 最后登录时间
	LastLoginIP string         `json:"last_login_ip,omitempty"`                         // 最后登录IP
	CreatedAt   time.Time      `json:"created_at" example:"2023-01-01T12:00:00Z"`       // 创建时间
	Roles       []RoleResponse `json:"roles,omitempty"`                                 // 用户角色
}

// UserListRequest 用户列表请求
type UserListRequest struct {
	Email  string `form:"email" example:"user@example.com"` // 用户邮箱，模糊查询
	Name   string `form:"name" example:"张"`                 // 用户姓名，模糊查询
	Status int    `form:"status" example:"1"`               // 状态：1-正常，2-禁用，3-锁定
	Page   int    `form:"page" example:"1"`                 // 页码
	Size   int    `form:"size" example:"10"`                // 每页数量
}

// UserListResponse 用户列表响应
type UserListResponse struct {
	Total int64          `json:"total" example:"100"` // 总数
	List  []UserResponse `json:"list"`                // 用户列表
}

// LoginResponse 登录响应
type LoginResponse struct {
	Token        string       `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`         // 访问令牌
	RefreshToken string       `json:"refresh_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."` // 刷新令牌
	ExpiresAt    int64        `json:"expires_at" example:"1672531200"`                                 // 过期时间戳
	UserInfo     UserResponse `json:"user_info"`                                                       // 用户信息
}
