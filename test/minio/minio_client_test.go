package minio_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"oss-backend/internal/interfaces"
	"oss-backend/mocks"
	pkgminio "oss-backend/pkg/minio"
)

// 创建类型别名以确保类型安全
type ObjectInfoChan <-chan minio.ObjectInfo

// 使用真实客户端的测试
func TestMinioConnection_WithRealClient(t *testing.T) {
	// 跳过这个测试，因为它需要真实的MinIO服务器
	t.Skip("跳过需要真实MinIO服务器的测试")

	// 加载配置
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("../../configs")
	err := viper.ReadInConfig()
	if err != nil {
		t.Fatalf("无法读取配置文件: %v", err)
	}

	// 打印配置信息
	endpoint := viper.GetString("minio.endpoint")
	accessKey := viper.GetString("minio.access_key")
	secretKey := viper.GetString("minio.secret_key")
	useSSL := viper.GetBool("minio.use_ssl")
	bucketName := viper.GetString("storage.bucket")
	if bucketName == "" {
		bucketName = "oss"
	}

	t.Logf("MinIO配置：")
	t.Logf("Endpoint: %s", endpoint)
	t.Logf("AccessKey: %s", accessKey)
	t.Logf("UseSSL: %v", useSSL)
	t.Logf("Bucket: %s", bucketName)

	// 初始化MinIO客户端
	config := pkgminio.Config{
		Endpoint:  endpoint,
		AccessKey: accessKey,
		SecretKey: secretKey,
		UseSSL:    useSSL,
	}

	// 创建客户端
	client, err := pkgminio.NewClient(config)
	if err != nil {
		t.Fatalf("初始化MinIO客户端失败: %v", err)
	}

	// 运行通用测试
	runMinioTest(t, client, bucketName)
}

// 使用Mock客户端的测试
func TestMinioConnection_WithMockClient(t *testing.T) {
	// 创建Mock客户端
	mockClient := new(mocks.MinioClient)
	bucketName := "test-bucket"
	testFileName := fmt.Sprintf("test-file-%d.txt", time.Now().Unix())
	testContent := []byte("这是一个MinIO连接测试文件，创建于 " + time.Now().Format("2006-01-02 15:04:05"))

	// 模拟存储桶列表
	mockBuckets := []minio.BucketInfo{
		{Name: bucketName, CreationDate: time.Now()},
	}

	// 模拟对象信息
	mockObjectInfo := minio.ObjectInfo{
		Key:          testFileName,
		Size:         int64(len(testContent)),
		LastModified: time.Now(),
		ContentType:  "text/plain",
	}

	// 创建通道处理函数
	createEmptyObjectsChan := func(ctx context.Context, bucket, prefix string, recursive bool) <-chan minio.ObjectInfo {
		ch := make(chan minio.ObjectInfo)
		go func() {
			close(ch)
		}()
		return ch
	}

	createObjectsChanWithItems := func(ctx context.Context, bucket, prefix string, recursive bool) <-chan minio.ObjectInfo {
		ch := make(chan minio.ObjectInfo, 1)
		go func() {
			ch <- mockObjectInfo
			close(ch)
		}()
		return ch
	}

	// 设置mock行为
	// 1. 列出桶
	mockClient.On("ListBuckets", mock.Anything).Return(mockBuckets, nil)

	// 2. 检查桶是否存在
	mockClient.On("BucketExists", mock.Anything, bucketName).Return(true, nil)

	// 3. 列出对象 - 第一次调用返回空
	mockClient.On("ListObjects", mock.Anything, bucketName, "", true).
		Return(createEmptyObjectsChan).Once()

	// 4. 上传对象
	mockClient.On("PutObject", mock.Anything, bucketName, mock.AnythingOfType("string"),
		mock.AnythingOfType("*bytes.Reader"), mock.AnythingOfType("int64"),
		mock.AnythingOfType("string")).Return(nil)

	// 5. 获取对象状态
	mockClient.On("StatObject", mock.Anything, bucketName, mock.AnythingOfType("string"),
		mock.Anything).Return(mockObjectInfo, nil)

	// 6. 获取对象内容
	mockClient.On("GetObject", mock.Anything, bucketName, mock.AnythingOfType("string"),
		mock.Anything).Return(io.NopCloser(bytes.NewReader(testContent)), nil)

	// 7. 列出对象 - 第二次调用返回有对象
	mockClient.On("ListObjects", mock.Anything, bucketName, "", true).
		Return(createObjectsChanWithItems).Once()

	// 运行测试
	runMinioTest(t, mockClient, bucketName)
}

// 通用的MinIO测试流程，可以传入真实客户端或Mock客户端
func runMinioTest(t *testing.T, client interfaces.MinioClient, bucketName string) {
	ctx := context.Background()

	// 1. 列出所有桶
	t.Log("测试1: 列出所有桶")
	buckets, err := client.ListBuckets(ctx)
	assert.NoError(t, err, "列出桶失败")

	if len(buckets) == 0 {
		t.Log("警告: 没有找到任何桶")
	} else {
		t.Logf("找到 %d 个桶:", len(buckets))
		for i, bucket := range buckets {
			t.Logf("  %d. %s (创建于: %s)", i+1, bucket.Name, bucket.CreationDate.Format("2006-01-02 15:04:05"))
		}
	}

	// 2. 检查指定的桶是否存在
	t.Logf("测试2: 检查桶 '%s' 是否存在", bucketName)
	exists, err := client.BucketExists(ctx, bucketName)
	assert.NoError(t, err, "检查桶是否存在失败")

	if !exists {
		t.Logf("桶 '%s' 不存在，尝试创建...", bucketName)
		err = client.MakeBucket(ctx, bucketName)
		assert.NoError(t, err, "创建桶失败")
		t.Logf("成功创建桶 '%s'", bucketName)
	} else {
		t.Logf("桶 '%s' 已存在", bucketName)
	}

	// 3. 列出桶中的所有对象
	t.Logf("测试3: 列出桶 '%s' 中的所有对象", bucketName)
	objects := client.ListObjects(ctx, bucketName, "", true)

	objectCount := 0
	for object := range objects {
		if object.Err != nil {
			t.Logf("获取对象信息出错: %v", object.Err)
			continue
		}
		objectCount++
		t.Logf("  - 对象: %s (大小: %d, 最后修改: %s)",
			object.Key, object.Size, object.LastModified.Format("2006-01-02 15:04:05"))
	}

	if objectCount == 0 {
		t.Logf("警告: 桶 '%s' 中没有找到任何对象", bucketName)
	} else {
		t.Logf("桶 '%s' 中共有 %d 个对象", bucketName, objectCount)
	}

	// 4. 尝试上传一个测试文件
	t.Log("测试4: 尝试上传测试文件")
	testFileName := fmt.Sprintf("test-file-%d.txt", time.Now().Unix())
	testContent := []byte("这是一个MinIO连接测试文件，创建于 " + time.Now().Format("2006-01-02 15:04:05"))

	// 上传测试文件
	err = client.PutObject(ctx, bucketName, testFileName, bytes.NewReader(testContent), int64(len(testContent)), "text/plain")
	assert.NoError(t, err, "上传测试文件失败")
	t.Logf("成功上传测试文件: %s", testFileName)

	// 5. 确认文件上传成功
	t.Log("测试5: 验证测试文件上传成功")
	stat, err := client.StatObject(ctx, bucketName, testFileName, nil)
	assert.NoError(t, err, "获取测试文件状态失败")
	t.Logf("文件详情: 名称=%s, 大小=%d, 类型=%s", stat.Key, stat.Size, stat.ContentType)

	// 6. 下载测试文件以验证内容
	t.Log("测试6: 下载测试文件验证内容")
	obj, err := client.GetObject(ctx, bucketName, testFileName, nil)
	assert.NoError(t, err, "下载测试文件失败")
	defer obj.Close()

	downloadedContent := make([]byte, stat.Size)
	_, err = obj.Read(downloadedContent)
	assert.NoError(t, err, "读取下载的文件内容失败")

	// 在mock测试中，我们总是返回相同的内容，所以这个比较可能不准确
	// 我们可以检查长度或其他属性来确认
	if _, ok := client.(*mocks.MinioClient); ok {
		t.Log("使用Mock客户端，跳过内容比较")
	} else if string(downloadedContent) == string(testContent) {
		t.Log("验证成功: 下载的内容与上传的内容匹配")
	} else {
		t.Errorf("验证失败: 下载的内容与上传的内容不匹配")
		t.Logf("原始内容: %s", string(testContent))
		t.Logf("下载内容: %s", string(downloadedContent))
	}

	// 7. 列出桶中的对象确认是否包含我们刚上传的文件
	t.Log("测试7: 再次列出桶中的对象，验证是否包含测试文件")
	objects = client.ListObjects(ctx, bucketName, "", true)

	found := false
	for object := range objects {
		if object.Err != nil {
			continue
		}
		t.Logf("  - 对象: %s (大小: %d, 类型: %s, 最后修改: %s)",
			object.Key, object.Size, getContentType(object.Key), object.LastModified.Format("2006-01-02 15:04:05"))
		if strings.Contains(object.Key, "test-file-") {
			found = true
		}
	}

	if found {
		t.Logf("测试文件在桶中已找到")
	} else {
		t.Errorf("测试文件在桶中未找到，上传可能失败")
	}

	// 测试结论
	t.Log("MinIO测试完成，观察日志判断问题所在")
}

// getContentType 根据文件名获取内容类型
func getContentType(filename string) string {
	ext := strings.ToLower(filename[strings.LastIndex(filename, ".")+1:])
	switch ext {
	case "jpg", "jpeg":
		return "image/jpeg"
	case "png":
		return "image/png"
	case "gif":
		return "image/gif"
	case "pdf":
		return "application/pdf"
	case "txt":
		return "text/plain"
	case "html", "htm":
		return "text/html"
	case "css":
		return "text/css"
	case "js":
		return "application/javascript"
	case "json":
		return "application/json"
	case "xml":
		return "application/xml"
	case "zip":
		return "application/zip"
	case "tar":
		return "application/x-tar"
	default:
		return "application/octet-stream"
	}
}
