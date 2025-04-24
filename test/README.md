# OSS-Backend 自动化测试工具

这是一个基于Swagger文档自动生成动态测试用例的工具，能够实现API的自动化测试，支持随机数据生成。

## 功能特点

- 基于Swagger文档自动生成测试用例
- 支持API接口的自动测试
- 自动生成随机测试数据
- 支持测试用例之间的依赖关系
- 支持变量存储和传递
- 支持文件上传测试
- 生成详细的测试报告

## 快速开始

### 构建测试工具

```bash
cd test
go build -o oss_tester
```

### 生成测试配置

```bash
# 从Swagger文档生成测试配置
./oss_tester -swagger ../docs/swagger/swagger.json -output test_config.json
```

### 运行测试

```bash
# 运行所有测试用例
./oss_tester -run
```

### 其他命令选项

```bash
# 显示帮助信息
./oss_tester -help

# 指定API基础URL
./oss_tester -base http://localhost:8080

# 生成配置并直接运行
./oss_tester -swagger ../docs/swagger/swagger.json -output test_config.json -run
```

## 测试配置文件结构

测试配置文件是一个JSON文件，包含以下主要部分：

```json
{
  "base_url": "http://localhost:8080",
  "test_cases": [
    {
      "name": "测试用例名称",
      "description": "测试用例描述",
      "skip": false,
      "depends_on": ["依赖的测试用例"],
      "api": {
        "endpoint": "/api/endpoint",
        "method": "POST",
        "auth": true
      },
      "request": {
        "param1": "value1",
        "param2": "value2"
      },
      "file_upload": {
        "field_name": "file",
        "file_name": "test.txt",
        "file_path": "test_file.txt"
      },
      "expect": {
        "status_code": 200,
        "contains": ["期望包含的字符串"],
        "not_contains": ["期望不包含的字符串"],
        "json": {
          "code": 200,
          "data.field": "期望值"
        }
      },
      "store": {
        "变量名": "data.field"
      }
    }
  ]
}
```

## 变量引用

在测试用例中，可以使用`${变量名}`的形式引用之前测试中存储的变量：

```json
{
  "request": {
    "user_id": "${USER_ID}"
  }
}
```

## 支持的测试场景

目前支持的测试场景包括：

1. 用户注册和登录
2. 用户信息获取和修改
3. 群组管理
4. 项目管理
5. 文件管理
6. 权限管理

## 疑难解答

如遇到问题，请检查：

1. 后端服务是否正常运行
2. Swagger文档是否完整有效
3. 测试配置文件格式是否正确
4. 网络连接是否畅通 