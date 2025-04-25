package minio_test

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	pkgminio "oss-backend/pkg/minio"
)

// TestGetObjectName 测试生成对象名称的函数
func TestGetObjectName(t *testing.T) {
	// 测试用例
	testCases := []struct {
		name      string
		projectID string
		filePath  string
		fileName  string
		expected  string
	}{
		{
			name:      "普通文件路径",
			projectID: "123",
			filePath:  "documents",
			fileName:  "test.txt",
			expected:  "project_123/documents/test.txt",
		},
		{
			name:      "带有前导斜杠的路径",
			projectID: "456",
			filePath:  "/photos",
			fileName:  "image.jpg",
			expected:  "project_456/photos/image.jpg",
		},
		{
			name:      "多级路径",
			projectID: "789",
			filePath:  "/data/uploads/2023",
			fileName:  "report.pdf",
			expected:  "project_789/data/uploads/2023/report.pdf",
		},
		{
			name:      "空路径",
			projectID: "abc",
			filePath:  "",
			fileName:  "empty-path.txt",
			expected:  "project_abc/empty-path.txt",
		},
	}

	// 执行测试
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := pkgminio.GetObjectName(tc.projectID, tc.filePath, tc.fileName)
			// 统一转换为Unix风格的路径分隔符进行比较
			normalizedResult := strings.ReplaceAll(result, "\\", "/")
			assert.Equal(t, tc.expected, normalizedResult, "对象名称不匹配")
		})
	}
}

// TestBucketOperations 测试桶操作辅助函数
func TestBucketOperations(t *testing.T) {
	// 跳过集成测试，因为它需要真实的MinIO服务器
	t.Skip("跳过需要真实MinIO服务器的集成测试")

	// 配置信息（如果需要执行测试，请填写有效的信息）
	config := pkgminio.Config{
		Endpoint:  "47.96.113.223:9000",
		AccessKey: "minioadmin",
		SecretKey: "minioadmin",
		UseSSL:    false,
	}

	// 创建客户端
	client, err := pkgminio.NewClient(config)
	if err != nil {
		t.Fatalf("创建MinIO客户端失败: %v", err)
	}

	// 测试创建桶
	ctx := context.Background()
	testBucket := "test-bucket-operations"

	t.Run("CreateBucketIfNotExists", func(t *testing.T) {
		err := client.CreateBucketIfNotExists(ctx, testBucket)
		assert.NoError(t, err, "创建桶失败")

		// 再次调用应该不会出错
		err = client.CreateBucketIfNotExists(ctx, testBucket)
		assert.NoError(t, err, "重复创建桶应该不返回错误")
	})

	t.Run("BucketExists", func(t *testing.T) {
		exists, err := client.BucketExists(ctx, testBucket)
		assert.NoError(t, err, "检查桶是否存在失败")
		assert.True(t, exists, "桶应该存在")

		exists, err = client.BucketExists(ctx, "non-existent-bucket-"+randomString(8))
		assert.NoError(t, err, "检查不存在的桶不应出错")
		assert.False(t, exists, "不存在的桶应返回false")
	})
}

// 帮助函数：生成随机字符串
func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, n)
	for i := 0; i < n; i++ {
		result[i] = letters[i%len(letters)]
	}
	return string(result)
}
