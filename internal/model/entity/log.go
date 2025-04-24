package entity

import (
	"time"

	"gorm.io/gorm"
)

// OperationLog 操作日志模型
type OperationLog struct {
	ID         string         `gorm:"primaryKey;type:bigint unsigned" json:"id"`
	UserID     string         `gorm:"type:bigint unsigned;index;not null" json:"user_id"`
	Action     string         `gorm:"type:varchar(50);not null" json:"action"` // 操作类型: create, update, delete, share, etc.
	Module     string         `gorm:"type:varchar(50);not null" json:"module"` // 模块: file, group, project, etc.
	TargetID   string         `gorm:"type:bigint unsigned;index" json:"target_id"`
	TargetType string         `gorm:"type:varchar(50)" json:"target_type"` // 目标类型: file, folder, group, project, etc.
	Details    string         `gorm:"type:text" json:"details"`            // 详细信息，JSON格式
	IP         string         `gorm:"type:varchar(50)" json:"ip"`          // 操作IP
	CreatedAt  time.Time      `gorm:"index" json:"created_at"`             // 操作时间
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`

	User User `gorm:"foreignKey:UserID" json:"user"`
}

// TableName 表名
func (OperationLog) TableName() string {
	return "operation_logs"
}

// AccessLog 访问日志模型
type AccessLog struct {
	ID         string         `gorm:"primaryKey;type:bigint unsigned" json:"id"`
	UserID     string         `gorm:"type:bigint unsigned;index" json:"user_id"` // 可能为空，表示匿名访问
	Method     string         `gorm:"type:varchar(10);not null" json:"method"`   // HTTP方法: GET, POST, PUT, DELETE, etc.
	Path       string         `gorm:"type:varchar(255);not null" json:"path"`    // 请求路径
	StatusCode int            `gorm:"type:int;not null" json:"status_code"`      // 响应状态码
	IP         string         `gorm:"type:varchar(50);not null" json:"ip"`       // 访问IP
	UserAgent  string         `gorm:"type:varchar(255)" json:"user_agent"`       // 用户代理
	Duration   int64          `gorm:"not null" json:"duration"`                  // 请求耗时(毫秒)
	CreatedAt  time.Time      `gorm:"index" json:"created_at"`                   // 访问时间
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`

	User User `gorm:"foreignKey:UserID" json:"user"`
}

// TableName 表名
func (AccessLog) TableName() string {
	return "access_logs"
}

// Log 操作日志模型
type Log struct {
	ID              string    `gorm:"primaryKey;type:bigint unsigned" json:"id"`
	UserID          string    `gorm:"type:bigint unsigned;not null;index:idx_user_time,priority:1" json:"user_id"`
	GroupID         string    `gorm:"type:bigint unsigned;not null" json:"group_id"`
	ProjectID       string    `gorm:"type:bigint unsigned;not null;index:idx_project_time,priority:1" json:"project_id"`
	FileID          *string   `gorm:"type:bigint unsigned" json:"file_id"`
	Operation       string    `gorm:"type:varchar(20);not null" json:"operation"`
	IPAddress       string    `gorm:"type:varchar(50);not null" json:"ip_address"`
	UserAgent       string    `gorm:"type:varchar(255)" json:"user_agent"`
	Status          int       `gorm:"default:200;not null" json:"status"`
	CreatedAt       time.Time `gorm:"not null;index:idx_user_time,priority:2;index:idx_project_time,priority:2" json:"created_at"`
	RequestDetails  string    `gorm:"type:text" json:"request_details"`
	ResponseDetails string    `gorm:"type:text" json:"response_details"`
	ExecutionTime   int       `json:"execution_time"` // 毫秒

	User    User    `gorm:"foreignKey:UserID" json:"user"`
	Group   Group   `gorm:"foreignKey:GroupID" json:"group"`
	Project Project `gorm:"foreignKey:ProjectID" json:"project"`
	File    *File   `gorm:"foreignKey:FileID" json:"file,omitempty"`
}

// TableName 表名
func (Log) TableName() string {
	return "logs"
}

// StorageStat 存储统计模型
type StorageStat struct {
	ID           string    `gorm:"primaryKey;type:bigint unsigned" json:"id"`
	GroupID      string    `gorm:"type:bigint unsigned;not null" json:"group_id"`
	ProjectID    string    `gorm:"type:bigint unsigned;not null;index:idx_project_date,priority:1" json:"project_id"`
	StatDate     time.Time `gorm:"not null;index:idx_project_date,priority:2;index:idx_date" json:"stat_date"`
	FileCount    int64     `gorm:"default:0;not null" json:"file_count"`
	TotalSize    int64     `gorm:"default:0;not null" json:"total_size"`
	IncreaseSize int64     `gorm:"default:0;not null" json:"increase_size"`
	CreatedAt    time.Time `json:"created_at"`

	Group   Group   `gorm:"foreignKey:GroupID" json:"group"`
	Project Project `gorm:"foreignKey:ProjectID" json:"project"`
}

// TableName 表名
func (StorageStat) TableName() string {
	return "storage_stats"
}

// 操作类型常量
const (
	OperationLogin        = "login"
	OperationLogout       = "logout"
	OperationUpload       = "upload"
	OperationDownload     = "download"
	OperationDelete       = "delete"
	OperationRestore      = "restore"
	OperationShare        = "share"
	OperationCancelShare  = "cancel_share"
	OperationCreateFolder = "create_folder"
	OperationRename       = "rename"
	OperationMove         = "move"
	OperationCopy         = "copy"
)
