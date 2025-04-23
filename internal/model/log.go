package model

import (
	"time"
)

// Log 操作日志模型
type Log struct {
	ID              uint64    `gorm:"primaryKey;type:bigint unsigned" json:"id"`
	UserID          uint64    `gorm:"type:bigint unsigned;not null;index:idx_user_time,priority:1" json:"user_id"`
	GroupID         uint64    `gorm:"type:bigint unsigned;not null" json:"group_id"`
	ProjectID       uint64    `gorm:"type:bigint unsigned;not null;index:idx_project_time,priority:1" json:"project_id"`
	FileID          *uint64   `gorm:"type:bigint unsigned" json:"file_id"`
	Operation       string    `gorm:"type:varchar(20);not null" json:"operation"`
	IPAddress       string    `gorm:"type:varchar(50);not null" json:"ip_address"`
	UserAgent       string    `gorm:"type:varchar(255)" json:"user_agent"`
	Status          int       `gorm:"default:200;not null" json:"status"`
	CreatedAt       time.Time `gorm:"not null;index:idx_user_time,priority:2;index:idx_project_time,priority:2" json:"created_at"`
	RequestDetails  string    `gorm:"type:text" json:"request_details"`
	ResponseDetails string    `gorm:"type:text" json:"response_details"`
	ExecutionTime   int       `json:"execution_time"` // 毫秒
	
	User    User     `gorm:"foreignKey:UserID" json:"user"`
	Group   Group    `gorm:"foreignKey:GroupID" json:"group"`
	Project Project  `gorm:"foreignKey:ProjectID" json:"project"`
	File    *File    `gorm:"foreignKey:FileID" json:"file,omitempty"`
}

// TableName 表名
func (Log) TableName() string {
	return "logs"
}

// StorageStat 存储统计模型
type StorageStat struct {
	ID           uint64    `gorm:"primaryKey;type:bigint unsigned" json:"id"`
	GroupID      uint64    `gorm:"type:bigint unsigned;not null" json:"group_id"`
	ProjectID    uint64    `gorm:"type:bigint unsigned;not null;index:idx_project_date,priority:1" json:"project_id"`
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