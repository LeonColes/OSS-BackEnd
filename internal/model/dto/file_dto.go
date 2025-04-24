package dto

import "time"

// ===== 请求结构 =====

// FileUploadRequest 文件上传请求
type FileUploadRequest struct {
	ProjectID uint64 `form:"project_id" binding:"required"` // 项目ID
	Path      string `form:"path" binding:"omitempty"`      // 上传路径，默认为根目录
	Comment   string `form:"comment" binding:"omitempty"`   // 文件注释
	Overwrite bool   `form:"overwrite" binding:"omitempty"` // 是否覆盖同名文件
}

// FileDownloadRequest 文件下载请求
type FileDownloadRequest struct {
	FileID uint64 `form:"file_id" binding:"required"` // 文件ID
}

// FileListRequest 文件列表请求
type FileListRequest struct {
	ProjectID      uint64 `form:"project_id" binding:"required"` // 项目ID
	Path           string `form:"path" binding:"omitempty"`      // 文件路径，默认为根目录
	Recursive      bool   `form:"recursive" binding:"omitempty"` // 是否递归获取子目录
	Page           int    `form:"page,default=1"`                // 页码
	Size           int    `form:"size,default=20"`               // 每页大小
	OrderBy        string `form:"order_by,default=updated_at"`   // 排序字段
	OrderDirection string `form:"order_direction,default=desc"`  // 排序方向
}

// FileFolderCreateRequest 创建文件夹请求
type FileFolderCreateRequest struct {
	ProjectID  uint64 `json:"project_id" binding:"required"`  // 项目ID
	Path       string `json:"path" binding:"omitempty"`       // 文件夹父路径
	FolderName string `json:"folder_name" binding:"required"` // 文件夹名称
}

// FileDeleteRequest 文件删除请求
type FileDeleteRequest struct {
	FileID uint64 `json:"file_id" binding:"required"` // 文件ID
}

// FileRestoreRequest 文件恢复请求
type FileRestoreRequest struct {
	FileID uint64 `json:"file_id" binding:"required"` // 文件ID
}

// FileShareCreateRequest 创建文件分享请求
type FileShareCreateRequest struct {
	FileID        uint64 `json:"file_id" binding:"required"`               // 文件ID
	Password      string `json:"password" binding:"omitempty"`             // 访问密码
	ExpireHours   int    `json:"expire_hours" binding:"omitempty"`         // 过期小时数，0表示永不过期
	DownloadLimit int    `json:"download_limit" binding:"omitempty,min=0"` // 下载次数限制，0表示无限制
}

// FileShareAccessRequest 访问分享文件请求
type FileShareAccessRequest struct {
	ShareCode string `json:"share_code" binding:"required"` // 分享码
	Password  string `json:"password" binding:"omitempty"`  // 访问密码
}

// ===== 响应结构 =====

// FileResponse 文件响应
type FileResponse struct {
	ID             uint64     `json:"id"`
	ProjectID      uint64     `json:"project_id"`
	FileName       string     `json:"file_name"`
	FilePath       string     `json:"file_path"`
	FullPath       string     `json:"full_path"`
	FileSize       int64      `json:"file_size"`
	MimeType       string     `json:"mime_type"`
	Extension      string     `json:"extension"`
	IsFolder       bool       `json:"is_folder"`
	IsDeleted      bool       `json:"is_deleted"`
	UploaderID     uint64     `json:"uploader_id"`
	UploaderName   string     `json:"uploader_name"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	DeletedAt      *time.Time `json:"deleted_at,omitempty"`
	DeletedBy      *uint64    `json:"deleted_by,omitempty"`
	DeleterName    string     `json:"deleter_name,omitempty"`
	CurrentVersion int        `json:"current_version"`
	PreviewURL     string     `json:"preview_url,omitempty"`
}

// FileVersionResponse 文件版本响应
type FileVersionResponse struct {
	ID           uint64    `json:"id"`
	FileID       uint64    `json:"file_id"`
	Version      int       `json:"version"`
	FileHash     string    `json:"file_hash"`
	FileSize     int64     `json:"file_size"`
	UploaderID   uint64    `json:"uploader_id"`
	UploaderName string    `json:"uploader_name"`
	CreatedAt    time.Time `json:"created_at"`
	Comment      string    `json:"comment"`
}

// FileShareResponse 文件分享响应
type FileShareResponse struct {
	ID            uint64     `json:"id"`
	FileID        uint64     `json:"file_id"`
	FileName      string     `json:"file_name"`
	FileSize      int64      `json:"file_size"`
	MimeType      string     `json:"mime_type"`
	ShareCode     string     `json:"share_code"`
	HasPassword   bool       `json:"has_password"`
	ExpireAt      *time.Time `json:"expire_at,omitempty"`
	DownloadLimit int        `json:"download_limit"`
	DownloadCount int        `json:"download_count"`
	CreatedAt     time.Time  `json:"created_at"`
	CreatorName   string     `json:"creator_name"`
}

// FileListResponse 文件列表响应
type FileListResponse struct {
	Total int64          `json:"total"`
	Items []FileResponse `json:"items"`
}

// FileVersionListResponse 文件版本列表响应
type FileVersionListResponse struct {
	FileID uint64                `json:"file_id"`
	Total  int                   `json:"total"`
	Items  []FileVersionResponse `json:"items"`
}
