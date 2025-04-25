package main

import (
	"context"
	"fmt"
	"log"
	"time"

	pkgminio "oss-backend/pkg/minio"

	"github.com/spf13/viper"
)

func main() {
	// 1. 初始化配置
	if err := initConfig(); err != nil {
		log.Fatalf("初始化配置失败: %v", err)
	}

	// 2. 读取 MinIO 配置
	minioConfig := pkgminio.Config{
		Endpoint:  viper.GetString("minio.endpoint"),
		AccessKey: viper.GetString("minio.access_key"),
		SecretKey: viper.GetString("minio.secret_key"),
		UseSSL:    viper.GetBool("minio.use_ssl"),
	}

	fmt.Println("MinIO 配置信息:")
	fmt.Printf("Endpoint: %s\n", minioConfig.Endpoint)
	fmt.Printf("AccessKey: %s\n", minioConfig.AccessKey)
	fmt.Printf("UseSSL: %v\n", minioConfig.UseSSL)

	// 3. 创建 MinIO 客户端
	minioClient, err := pkgminio.NewClient(minioConfig)
	if err != nil {
		log.Fatalf("初始化 MinIO 客户端失败: %v", err)
	}

	// 4. 创建上下文
	ctx := context.Background()

	// 5. 创建自定义桶名称
	bucketName := fmt.Sprintf("test-bucket-%d", time.Now().Unix())
	fmt.Printf("准备创建存储桶: %s\n", bucketName)

	// 6. 第一种方式：使用 CreateBucketIfNotExists 方法
	err = minioClient.CreateBucketIfNotExists(ctx, bucketName)
	if err != nil {
		log.Fatalf("使用 CreateBucketIfNotExists 创建存储桶失败: %v", err)
	}
	fmt.Printf("成功创建存储桶 (方式1): %s\n", bucketName)

	// 7. 第二种方式：使用 MakeBucket 方法
	bucketName2 := fmt.Sprintf("test-bucket2-%d", time.Now().Unix())
	fmt.Printf("准备创建存储桶: %s\n", bucketName2)

	// 检查桶是否已存在
	exists, err := minioClient.BucketExists(ctx, bucketName2)
	if err != nil {
		log.Fatalf("检查存储桶是否存在失败: %v", err)
	}

	if !exists {
		// 创建桶
		err = minioClient.MakeBucket(ctx, bucketName2)
		if err != nil {
			log.Fatalf("使用 MakeBucket 创建存储桶失败: %v", err)
		}
		fmt.Printf("成功创建存储桶 (方式2): %s\n", bucketName2)
	} else {
		fmt.Printf("存储桶已存在: %s\n", bucketName2)
	}

	// 8. 列出所有存储桶
	buckets, err := minioClient.ListBuckets(ctx)
	if err != nil {
		log.Fatalf("列出存储桶失败: %v", err)
	}

	fmt.Println("\n当前所有存储桶:")
	for i, bucket := range buckets {
		fmt.Printf("%d. %s (创建时间: %s)\n", i+1, bucket.Name, bucket.CreationDate.Format("2006-01-02 15:04:05"))
	}
}

// 初始化配置
func initConfig() error {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./configs")
	return viper.ReadInConfig()
}
