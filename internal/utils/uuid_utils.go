package utils

import (
	"github.com/google/uuid"
)

// GenerateUUID 生成UUID字符串
func GenerateUUID() string {
	return uuid.New().String()
}

// IsValidUUID 检查字符串是否为有效的UUID
func IsValidUUID(u string) bool {
	_, err := uuid.Parse(u)
	return err == nil
}

// GenerateUserID 生成用户ID
func GenerateUserID() string {
	return GenerateUUID()
}

// GenerateGroupID 生成群组ID
func GenerateGroupID() string {
	return GenerateUUID()
}

// GenerateProjectID 生成项目ID
func GenerateProjectID() string {
	return GenerateUUID()
}

// GenerateFileID 生成文件ID
func GenerateFileID() string {
	return GenerateUUID()
}

// GenerateRecordID 生成记录ID
func GenerateRecordID() string {
	return GenerateUUID()
}
