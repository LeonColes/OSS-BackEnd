# OSS系统自动化测试工具

这是一个基于配置驱动的API自动化测试工具，用于测试OSS后端系统的API功能。

## 功能特点

- 环境检测功能，确保测试环境正常
- 基于JSON配置的测试用例，易于扩展
- 支持测试用例之间的依赖关系
- 支持变量传递和引用
- 支持文件上传测试
- 详细的测试报告输出

## 快速开始

### 构建测试工具

```bash
cd test
go build -o oss_tester
```

### 运行测试

运行所有测试用例:

```bash
./oss_tester
```

运行特定测试用例:

```bash
./oss_tester -test login_user
```

跳过环境检测:

```bash
./oss_tester -skip-env
```

使用自定义配置文件:

```bash
./oss_tester -config my_config.json
```

## 配置文件说明

配置文件使用JSON格式，包含以下结构:

```json
{
  "base_url": "http://localhost:8080/api/oss",  // API基础URL
  "test_cases": [
    {
      "name": "test_name",                 // 测试用例名称
      "description": "测试描述",           // 测试用例描述
      "skip": false,                       // 是否跳过此测试
      "depends_on": ["other_test"],        // 依赖的其他测试用例
      "api": {
        "endpoint": "/path/to/api",        // API路径
        "method": "GET",                   // HTTP方法
        "auth": true                       // 是否需要认证
      },
      "request": {                         // 请求参数
        "key": "value",
        "ref_key": "${variable_name}"      // 引用之前存储的变量
      },
      "file_upload": {                     // 文件上传配置（可选）
        "field_name": "file",
        "file_name": "test.txt",
        "file_path": "./test.txt"
      },
      "expect": {                          // 断言配置
        "status_code": 200,                // 期望的状态码
        "contains": ["success"],           // 期望响应包含的字符串
        "not_contains": ["error"],         // 期望响应不包含的字符串
        "json": {                          // 期望的JSON响应
          "code": 0,
          "data.name": "value"             // 支持点标记法访问嵌套字段
        }
      },
      "store": {                           // 存储响应中的值（可选）
        "var_name": "data.field"           // 存储response.data.field到var_name变量
      }
    }
  ]
}
```

## 当前支持的测试场景

1. 用户注册与登录
2. 获取用户信息
3. 群组创建与管理
4. 项目创建与管理
5. 文件上传、下载与管理
6. 文件分享功能

## 故障排除

- 确保OSS后端服务正在运行
- 检查配置文件中的`base_url`是否正确
- 环境检测失败时，检查服务健康状态
- 对于文件上传测试，确保测试文件存在 