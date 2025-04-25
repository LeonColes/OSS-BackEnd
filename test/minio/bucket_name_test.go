package minio_test

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"oss-backend/pkg/utils"
)

// TestSanitizeBucketName 测试存储桶名称规范化
func TestSanitizeBucketName(t *testing.T) {
	// 创建文件服务实例
	fileService := &mockFileService{}

	// 测试表
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"小写字母和数字", "test123", "group-test123"},
		{"大写字母", "TEST", "group-test"},
		{"特殊字符", "test@#$%", "group-test----"},
		{"空格", "test space", "group-test-space"},
		{"中文", "测试", "group------"},
		{"很长的键", "abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyz", "group-abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcd"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := fileService.sanitizeBucketName(tt.input)
			if got != tt.expected {
				t.Errorf("sanitizeBucketName(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

// 实现测试用的mock文件服务
type mockFileService struct{}

// 公开的sanitizeBucketName方法，用于测试
func (s *mockFileService) sanitizeBucketName(key string) string {
	// 这个实现需要与service.fileService.sanitizeBucketName保持一致
	// 为测试目的直接复制实现
	// 1. 将所有字符转为小写
	lowerKey := strings.ToLower(key)
	// 2. 替换所有非法字符为连字符
	reg := regexp.MustCompile(`[^a-z0-9\-]`)
	sanitizedKey := reg.ReplaceAllString(lowerKey, "-")
	// 3. 确保不以连字符开头或结尾
	sanitizedKey = strings.Trim(sanitizedKey, "-")
	// 4. 如果长度不足，添加前缀
	if len(sanitizedKey) < 3 {
		sanitizedKey = fmt.Sprintf("grp-%s", sanitizedKey)
	}
	// 5. 如果长度过长，截断
	if len(sanitizedKey) > 60 {
		sanitizedKey = sanitizedKey[:60]
	}

	return fmt.Sprintf("group-%s", sanitizedKey)
}

func TestGenerateBucketName(t *testing.T) {
	testCases := []struct {
		name      string
		projectID string
		expected  string
	}{
		{
			name:      "普通项目ID",
			projectID: "12345",
			expected:  "project-12345",
		},
		{
			name:      "带特殊字符的项目ID",
			projectID: "project_123@45",
			expected:  "project-project-123-45",
		},
		{
			name:      "空项目ID",
			projectID: "",
			expected:  "project-",
		},
		{
			name:      "很长的项目ID",
			projectID: "this-is-a-very-long-project-id-that-exceeds-the-valid-bucket-name-length",
			expected:  "project-this-is-a-very-long-project-id-that-excee",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := utils.GenerateBucketName(tc.projectID)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestGenerateBucketNameMaxLength(t *testing.T) {
	// MinIO桶名称最大长度为63个字符
	const maxBucketNameLength = 63

	// 生成非常长的项目ID
	longProjectID := "this-is-an-extremely-long-project-id-that-definitely-exceeds-the-valid-bucket-name-length-limitation"

	// 生成桶名称
	bucketName := utils.GenerateBucketName(longProjectID)

	// 验证生成的桶名称不超过最大长度
	assert.LessOrEqual(t, len(bucketName), maxBucketNameLength,
		"生成的桶名称长度应小于或等于最大允许长度")
}
