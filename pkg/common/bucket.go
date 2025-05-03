package common

import (
	"regexp"
	"strings"
)

// 存储桶名称的最大长度限制
const maxBucketNameLength = 63

// GenerateBucketName 根据项目ID生成标准化的MinIO存储桶名称
// 格式为: project-{sanitized-project-id}
// 处理特殊字符、大小写和长度限制
func GenerateBucketName(projectID string) string {
	// 1. 将所有字符转为小写
	lowerKey := strings.ToLower(projectID)

	// 2. 替换所有非法字符为连字符(只保留a-z、0-9和连字符)
	reg := regexp.MustCompile(`[^a-z0-9\-]`)
	sanitizedKey := reg.ReplaceAllString(lowerKey, "-")

	// 3. 确保不以连字符开头或结尾
	sanitizedKey = strings.Trim(sanitizedKey, "-")

	// 4. 添加前缀
	result := "project-" + sanitizedKey

	// 5. 如果长度过长，截断
	if len(result) > maxBucketNameLength {
		result = result[:maxBucketNameLength]
	}

	return result
}
