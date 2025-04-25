# OSS-Backend 测试指南

本文档提供了如何运行和管理OSS-Backend项目测试的说明。

## 前置条件

测试前，请确保:

1. 已安装Go环境 (推荐1.23.0+)
2. 已安装mockery工具 (可通过`go install github.com/vektra/mockery/v2@latest`安装)
3. 已生成mock文件 (可通过`make mock`生成)

## 运行测试

### 运行所有测试

```bash
make test
```

这个命令会:
1. 生成mock文件
2. 运行所有单元测试
3. 运行集成测试(如果有)

### 运行单元测试

```bash
make unit-test
```

### 测试特定包

```bash
make test-pkg PKG=./test/minio
```

您可以替换`./test/minio`为任何包路径。

### 测试特定文件

```bash
make test-file FILE=./test/minio/bucket_name_test.go
```

您可以替换`./test/minio/bucket_name_test.go`为任何测试文件路径。

## 生成Mock文件

在进行单元测试前，您需要生成相应的mock文件:

```bash
make mock
```

该命令会在`./mocks`目录下生成所有接口的mock实现。

## 测试目录结构

- `test/minio/`: MinIO相关测试
- `test/file/`: 文件服务测试
- `test/project/`: 项目服务测试
- `test/rbac/`: 权限管理测试
- `test/storage/`: 存储服务测试
- `test/user/`: 用户服务测试

## 注意事项

1. 某些测试可能需要依赖特定配置或环境变量
2. 一些集成测试需要对应的服务(如MySQL、Redis、MinIO)运行
3. 部分测试(例如rbac和storage测试)当前被标记为跳过，因为它们依赖未实现的mock 