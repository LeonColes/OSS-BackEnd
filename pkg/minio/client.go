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
)

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
func (c *Client) PutObject(ctx context.Context, bucketName, objectName string, reader io.Reader, size int64, opts PutObjectOptions) error {
	// 转换为MinIO选项
	options := minio.PutObjectOptions{}
	if opts.ContentType != "" {
		options.ContentType = opts.ContentType
	}

	// 上传对象
	_, err := c.client.PutObject(ctx, bucketName, objectName, reader, size, options)
	return err
}

// GetObject 获取对象
func (c *Client) GetObject(ctx context.Context, bucketName, objectName string, opts GetObjectOptions) (io.ReadCloser, error) {
	// 转换为MinIO选项
	options := minio.GetObjectOptions{}

	// 获取对象
	obj, err := c.client.GetObject(ctx, bucketName, objectName, options)
	if err != nil {
		return nil, err
	}

	return obj, nil
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

// ListObjects 列出对象
func (c *Client) ListObjects(ctx context.Context, bucketName, prefix string, recursive bool) <-chan minio.ObjectInfo {
	return c.client.ListObjects(ctx, bucketName, minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: recursive,
	})
}

// CreateBucketIfNotExists 创建桶(如果不存在)
func (c *Client) CreateBucketIfNotExists(ctx context.Context, bucketName string) error {
	exists, err := c.client.BucketExists(ctx, bucketName)
	if err != nil {
		return err
	}

	if !exists {
		// 创建桶
		err = c.client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
		if err != nil {
			return err
		}
	}

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

// GetObjectName 生成对象名称
func GetObjectName(projectID uint64, filePath, fileName string) string {
	// 构建对象名称
	objectPath := filepath.Join(fmt.Sprintf("project_%d", projectID), filePath)
	objectPath = filepath.ToSlash(objectPath) // 转换为UNIX路径格式

	// 确保路径以/开头，不以/结尾
	if !strings.HasPrefix(objectPath, "/") {
		objectPath = "/" + objectPath
	}
	if strings.HasSuffix(objectPath, "/") {
		objectPath = objectPath[:len(objectPath)-1]
	}

	// 拼接文件名
	return strings.TrimPrefix(filepath.Join(objectPath, fileName), "/")
}
