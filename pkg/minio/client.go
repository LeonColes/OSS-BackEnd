package minio

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"oss-backend/internal/utils"
)

// 确保Client实现了MinioClient接口
var _ utils.MinioClient = (*Client)(nil)

// Config MinIO配置
type Config struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	UseSSL    bool
}

// Client MinIO客户端包装
type Client struct {
	client *minio.Client
}

// NewClient 创建新的MinIO客户端
func NewClient(cfg Config) (*Client, error) {
	// 创建MinIO客户端
	mc, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, err
	}

	return &Client{client: mc}, nil
}

// PutObjectOptions 上传对象选项
type PutObjectOptions struct {
	ContentType string
}

// GetObjectOptions 获取对象选项
type GetObjectOptions struct{}

// PutObject 上传对象
func (c *Client) PutObject(ctx context.Context, bucketName, objectName string, reader io.Reader, size int64, opts interface{}) error {
	var options minio.PutObjectOptions

	// 转换选项类型
	if putOpts, ok := opts.(PutObjectOptions); ok {
		if putOpts.ContentType != "" {
			options.ContentType = putOpts.ContentType
		}
	} else if contentType, ok := opts.(string); ok && contentType != "" {
		options.ContentType = contentType
	}

	// 上传对象
	_, err := c.client.PutObject(ctx, bucketName, objectName, reader, size, options)
	return err
}

// GetObject 获取对象
func (c *Client) GetObject(ctx context.Context, bucketName, objectName string, opts interface{}) (io.ReadCloser, error) {
	// 转换为MinIO选项
	options := minio.GetObjectOptions{}

	// 获取对象
	obj, err := c.client.GetObject(ctx, bucketName, objectName, options)
	if err != nil {
		return nil, err
	}

	return obj, nil
}

// ListBuckets 列出所有桶
func (c *Client) ListBuckets(ctx context.Context) ([]minio.BucketInfo, error) {
	return c.client.ListBuckets(ctx)
}

// MakeBucket 创建存储桶
func (c *Client) MakeBucket(ctx context.Context, bucketName string) error {
	return c.client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
}

// BucketExists 检查存储桶是否存在
func (c *Client) BucketExists(ctx context.Context, bucketName string) (bool, error) {
	return c.client.BucketExists(ctx, bucketName)
}

// RemoveObject 删除对象
func (c *Client) RemoveObject(ctx context.Context, bucketName, objectName string) error {
	return c.client.RemoveObject(ctx, bucketName, objectName, minio.RemoveObjectOptions{})
}

// StatObject 获取对象信息
func (c *Client) StatObject(ctx context.Context, bucketName, objectName string, opts interface{}) (minio.ObjectInfo, error) {
	options := minio.StatObjectOptions{}
	return c.client.StatObject(ctx, bucketName, objectName, options)
}

// ListObjects 列出对象
func (c *Client) ListObjects(ctx context.Context, bucketName, prefix string, recursive bool) <-chan minio.ObjectInfo {
	return c.client.ListObjects(ctx, bucketName, minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: recursive,
	})
}

// CreateBucketIfNotExists 如果存储桶不存在，则创建
func (c *Client) CreateBucketIfNotExists(ctx context.Context, bucketName string) error {
	// 检查存储桶是否存在
	exists, err := c.client.BucketExists(ctx, bucketName)
	if err != nil {
		return fmt.Errorf("检查存储桶是否存在失败: %w", err)
	}

	// 如果存储桶已存在，直接返回
	if exists {
		return nil
	}

	// 创建存储桶
	err = c.client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
	if err != nil {
		return fmt.Errorf("创建存储桶失败: %w", err)
	}

	// 设置存储桶策略 (可选)
	// 这里可以设置桶的访问策略，例如公共读取或私有访问
	// 本例中我们设置为私有访问

	return nil
}

// UploadFile 上传文件
func (c *Client) UploadFile(ctx context.Context, bucketName, objectName string, reader io.Reader, fileSize int64, contentType string) (string, error) {
	// 检查桶是否存在，不存在则创建
	err := c.CreateBucketIfNotExists(ctx, bucketName)
	if err != nil {
		return "", fmt.Errorf("确保桶存在失败: %w", err)
	}

	// 上传文件
	info, err := c.client.PutObject(ctx, bucketName, objectName, reader, fileSize, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", fmt.Errorf("上传文件失败: %w", err)
	}

	return info.ETag, nil
}

// DownloadFile 下载文件
func (c *Client) DownloadFile(ctx context.Context, bucketName, objectName string) (io.ReadCloser, int64, error) {
	// 获取对象信息
	objInfo, err := c.client.StatObject(ctx, bucketName, objectName, minio.StatObjectOptions{})
	if err != nil {
		return nil, 0, fmt.Errorf("获取文件信息失败: %w", err)
	}

	// 获取对象
	obj, err := c.client.GetObject(ctx, bucketName, objectName, minio.GetObjectOptions{})
	if err != nil {
		return nil, 0, fmt.Errorf("获取文件失败: %w", err)
	}

	return obj, objInfo.Size, nil
}

// DeleteFile 删除文件
func (c *Client) DeleteFile(ctx context.Context, bucketName, objectName string) error {
	err := c.client.RemoveObject(ctx, bucketName, objectName, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("删除文件失败: %w", err)
	}

	return nil
}

// ListFiles 列出文件
func (c *Client) ListFiles(ctx context.Context, bucketName, prefix string) ([]minio.ObjectInfo, error) {
	// 列出对象
	objectCh := c.client.ListObjects(ctx, bucketName, minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: true,
	})

	var objects []minio.ObjectInfo
	for object := range objectCh {
		if object.Err != nil {
			return nil, fmt.Errorf("列出文件失败: %w", object.Err)
		}
		objects = append(objects, object)
	}

	return objects, nil
}

// CreateFolder 创建文件夹
func (c *Client) CreateFolder(ctx context.Context, bucketName, folderPath string) error {
	// 确保路径以/结尾
	if !strings.HasSuffix(folderPath, "/") {
		folderPath = folderPath + "/"
	}

	// 上传一个空文件作为文件夹标识
	_, err := c.client.PutObject(ctx, bucketName, folderPath, strings.NewReader(""), 0, minio.PutObjectOptions{})
	if err != nil {
		return fmt.Errorf("创建文件夹失败: %w", err)
	}

	return nil
}

// GetFileHash 获取文件哈希值(SHA-256)
func (c *Client) GetFileHash(reader io.Reader) (string, error) {
	hash := sha256.New()
	if _, err := io.Copy(hash, reader); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

// FileExists 检查文件是否存在
func (c *Client) FileExists(ctx context.Context, bucketName, objectName string) (bool, error) {
	_, err := c.client.StatObject(ctx, bucketName, objectName, minio.StatObjectOptions{})
	if err != nil {
		// 检查是否是文件不存在错误
		if minio.ToErrorResponse(err).Code == "NoSuchKey" {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

// GeneratePreSignedURL 生成预签名URL
func (c *Client) GeneratePreSignedURL(ctx context.Context, bucketName, objectName string, expiry time.Duration) (string, error) {
	// 生成预签名URL
	presignedURL, err := c.client.PresignedGetObject(ctx, bucketName, objectName, expiry, nil)
	if err != nil {
		return "", fmt.Errorf("生成预签名URL失败: %w", err)
	}

	return presignedURL.String(), nil
}

// GetPublicDownloadURL 获取公共下载URL，使用7天的过期时间
func (c *Client) GetPublicDownloadURL(ctx context.Context, bucketName, objectName string) (string, error) {
	// 使用7天过期时间
	expiry := time.Hour * 24 * 7
	return c.GeneratePreSignedURL(ctx, bucketName, objectName, expiry)
}

// GetObjectName 生成对象名称
func GetObjectName(projectID string, filePath, fileName string) string {
	// 构建对象名称
	objectPath := filepath.Join(fmt.Sprintf("project_%s", projectID), filePath)
	objectPath = filepath.ToSlash(objectPath) // 转换为UNIX路径格式

	// 确保路径以/开头，不以/结尾
	if !strings.HasPrefix(objectPath, "/") {
		objectPath = "/" + objectPath
	}
	objectPath = strings.TrimSuffix(objectPath, "/")

	// 拼接文件名
	return strings.TrimPrefix(filepath.Join(objectPath, fileName), "/")
}
