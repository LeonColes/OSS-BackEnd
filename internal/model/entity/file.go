package entity

import (
	"time"

	"gorm.io/gorm"
)

// File 文件模型
type File struct {
	ID             string         `gorm:"primaryKey;type:varchar(36)" json:"id"`
	ProjectID      string         `gorm:"type:varchar(36);not null;index:idx_project_path,priority:1" json:"project_id"`
	FileName       string         `gorm:"type:varchar(255);not null" json:"file_name"`
	FilePath       string         `gorm:"type:varchar(512);not null;index" json:"file_path"`
	FullPath       string         `gorm:"type:varchar(768);not null" json:"full_path"`
	FileHash       string         `gorm:"type:varchar(64);not null;index" json:"file_hash"`
	FileSize       int64          `gorm:"not null" json:"file_size"`
	MimeType       string         `gorm:"type:varchar(128)" json:"mime_type"`
	Extension      string         `gorm:"type:varchar(20)" json:"extension"`
	IsFolder       bool           `gorm:"default:false;not null" json:"is_folder"`
	IsDeleted      bool           `gorm:"default:false;not null;index" json:"is_deleted"`
	UploaderID     string         `gorm:"type:varchar(36);not null" json:"uploader_id"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      *time.Time     `json:"deleted_at"`
	DeletedBy      *string        `gorm:"type:varchar(36)" json:"deleted_by"`
	CurrentVersion int            `gorm:"default:1;not null" json:"current_version"`
	PreviewURL     string         `gorm:"type:varchar(512)" json:"preview_url"`
	GormDeletedAt  gorm.DeletedAt `gorm:"index" json:"-"` // 用于GORM的软删除，区别于业务上的IsDeleted标志

	Project  Project `gorm:"foreignKey:ProjectID" json:"project"`
	Uploader User    `gorm:"foreignKey:UploaderID" json:"uploader"`
	Deleter  *User   `gorm:"foreignKey:DeletedBy" json:"deleter,omitempty"`
}

// TableName 表名
func (File) TableName() string {
	return "files"
}

// FileVersion 文件版本模型
type FileVersion struct {
	ID         string    `gorm:"primaryKey;type:varchar(36)" json:"id"`
	FileID     string    `gorm:"type:varchar(36);not null;index:idx_file_version,priority:1" json:"file_id"`
	Version    int       `gorm:"not null;index:idx_file_version,priority:2" json:"version"`
	FileHash   string    `gorm:"type:varchar(64);not null" json:"file_hash"`
	FileSize   int64     `gorm:"not null" json:"file_size"`
	UploaderID string    `gorm:"type:varchar(36);not null" json:"uploader_id"`
	CreatedAt  time.Time `json:"created_at"`
	Comment    string    `gorm:"type:varchar(255)" json:"comment"`

	File     File `gorm:"foreignKey:FileID" json:"file"`
	Uploader User `gorm:"foreignKey:UploaderID" json:"uploader"`
}

// TableName 表名
func (FileVersion) TableName() string {
	return "file_versions"
}

// FileShare 文件分享模型
type FileShare struct {
	ID            string     `gorm:"primaryKey;type:varchar(36)" json:"id"`
	FileID        string     `gorm:"type:varchar(36);not null" json:"file_id"`
	UserID        string     `gorm:"type:varchar(36);not null;index" json:"user_id"`
	ShareCode     string     `gorm:"type:varchar(32);uniqueIndex;not null" json:"share_code"`
	Password      string     `gorm:"type:varchar(32)" json:"password,omitempty"`
	ExpireAt      *time.Time `json:"expire_at"`
	DownloadLimit int        `gorm:"default:0" json:"download_limit"` // 0表示无限制
	DownloadCount int        `gorm:"default:0" json:"download_count"`
	CreatedAt     time.Time  `json:"created_at"`

	File File `gorm:"foreignKey:FileID" json:"file"`
	User User `gorm:"foreignKey:UserID" json:"user"`
}

// TableName 表名
func (FileShare) TableName() string {
	return "file_shares"
}
