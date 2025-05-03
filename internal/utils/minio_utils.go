package utils

import (
	"context"
	"io"
	"time"

	miniolib "github.com/minio/minio-go/v7"
)

// MinioClient 接口定义MinIO客户端的能力
type MinioClient interface {
	// 桶操作
	ListBuckets(ctx context.Context) ([]miniolib.BucketInfo, error)
	BucketExists(ctx context.Context, bucketName string) (bool, error)
	MakeBucket(ctx context.Context, bucketName string) error

	// 对象操作
	ListObjects(ctx context.Context, bucketName, prefix string, recursive bool) <-chan miniolib.ObjectInfo
	PutObject(ctx context.Context, bucketName, objectName string, reader io.Reader, size int64, opts interface{}) error
	GetObject(ctx context.Context, bucketName, objectName string, opts interface{}) (io.ReadCloser, error)
	StatObject(ctx context.Context, bucketName, objectName string, opts interface{}) (miniolib.ObjectInfo, error)
	RemoveObject(ctx context.Context, bucketName, objectName string) error

	// 辅助功能
	CreateBucketIfNotExists(ctx context.Context, bucketName string) error
	UploadFile(ctx context.Context, bucketName, objectName string, reader io.Reader, fileSize int64, contentType string) (string, error)
	DownloadFile(ctx context.Context, bucketName, objectName string) (io.ReadCloser, int64, error)
	DeleteFile(ctx context.Context, bucketName, objectName string) error
	ListFiles(ctx context.Context, bucketName, prefix string) ([]miniolib.ObjectInfo, error)
	CreateFolder(ctx context.Context, bucketName, folderPath string) error
	GetFileHash(reader io.Reader) (string, error)
	FileExists(ctx context.Context, bucketName, objectName string) (bool, error)
	GeneratePreSignedURL(ctx context.Context, bucketName, objectName string, expiry time.Duration) (string, error)
	GetPublicDownloadURL(ctx context.Context, bucketName, objectName string) (string, error)
}
