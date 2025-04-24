# 贡献指南

感谢您对本项目的关注！这份贡献指南将帮助您了解如何参与到本项目的开发中。

## 目录

- [行为准则](#行为准则)
- [开发流程](#开发流程)
- [代码风格](#代码风格)
- [提交规范](#提交规范)
- [Pull Request流程](#pull-request流程)
- [Code Review标准](#code-review标准)
- [项目结构](#项目结构)
- [开发环境设置](#开发环境设置)

## 行为准则

参与本项目的所有贡献者都需要尊重项目的行为准则：

- 使用友善和包容的语言
- 尊重不同的观点和经验
- 优雅地接受建设性的批评
- 关注对社区最有利的事情
- 对其他社区成员表示同情和友善

## 开发流程

我们采用GitHub Flow工作流程:

1. 从`main`分支创建功能分支
2. 在功能分支上进行开发
3. 提交代码时遵循[提交规范](#提交规范)
4. 提交Pull Request到`main`分支
5. 等待Code Review并处理反馈意见
6. 合并后删除功能分支

### 分支命名规范

- 功能开发分支: `feature/功能名称`
- Bug修复分支: `fix/bug描述`
- 文档相关分支: `docs/文档描述`
- 性能优化分支: `perf/优化描述`
- 重构相关分支: `refactor/重构描述`

## 代码风格

我们使用自动化工具来保证代码风格的一致性：

### Go代码规范

- 遵循[Effective Go](https://golang.org/doc/effective_go)和[Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- 使用`gofmt`格式化代码
- 使用`golangci-lint`进行代码质量检查
- 测试覆盖率要求达到70%以上

### 命名规范

- 包名：使用小写字母，不使用下划线或混合大小写
- 接口名：通常以`er`结尾，如`Reader`, `Writer`
- 变量和函数名：使用驼峰命名法，如`userID`, `GetUserByID`
- 常量：使用全大写加下划线，如`MAX_CONNECTIONS`
- 私有成员：小写字母开头
- 公共成员：大写字母开头

### 注释规范

- 所有导出的函数、类型、变量和常量必须有注释
- 使用完整的句子，以被描述的事物名称开头
- 遵循[Godoc规范](https://blog.golang.org/godoc-documenting-go-code)

## 提交规范

我们使用[约定式提交](https://www.conventionalcommits.org/zh-hans/v1.0.0/)规范:

```
<类型>[可选的作用域]: <描述>

[可选的正文]

[可选的页脚]
```

### 提交类型

- `feat`: 新增功能
- `fix`: 修复Bug
- `docs`: 文档变更
- `style`: 代码格式调整
- `refactor`: 代码重构
- `perf`: 性能优化
- `test`: 测试相关变更
- `chore`: 构建过程或辅助工具的变更

### 示例

```
feat(user): 添加用户注册功能

实现了用户注册API和邮箱验证功能

Close #123
```

## Pull Request流程

1. 确保你的PR只包含一个具体的功能或修复
2. 更新相关文档
3. 确保所有测试通过
4. 填写PR模板，包括:
   - 相关Issue链接
   - 变更描述
   - 测试方法
   - 截图(如适用)

### PR标题格式

PR标题应遵循提交规范:

```
<类型>[可选的作用域]: <描述>
```

例如:
```
feat(auth): 添加OAuth2认证支持
```

## Code Review标准

审查者会关注以下几点:

- 代码质量和风格
- 测试覆盖率
- 文档完整性
- 性能影响
- 安全风险
- 兼容性考虑

## 项目结构

项目采用清晰的分层架构，请遵循现有的项目结构:

```
├── api/            # API接口定义
├── cmd/            # 主程序入口
├── configs/        # 配置文件
├── docs/           # 文档
├── internal/       # 内部代码，不对外暴露
│   ├── app/        # 应用层
│   ├── domain/     # 领域层
│   ├── infra/      # 基础设施层
│   └── interfaces/ # 接口层
├── pkg/            # 可复用的公共包
├── scripts/        # 脚本文件
└── test/           # 测试相关
```

开发新功能时，请遵循现有的分层和模块划分。

## 开发环境设置

### 前置条件

- Go 1.21+
- Docker & Docker Compose
- Make

### 本地开发

1. Fork本仓库
2. 克隆你的Fork到本地
3. 安装依赖
   ```
   go mod download
   ```
4. 使用docker-compose启动依赖服务
   ```
   docker-compose -f docker-compose.dev.yml up -d
   ```
5. 运行测试确保开发环境正常
   ```
   make test
   ```
6. 启动应用
   ```
   make run
   ```

### 常用开发命令

| 命令 | 描述 |
|------|------|
| `make run` | 启动应用 |
| `make test` | 运行测试 |
| `make lint` | 运行代码检查 |
| `make mock` | 生成Mock代码 |
| `make swagger` | 生成Swagger文档 |
| `make build` | 构建应用 |

---

感谢您的贡献！如有任何问题，请在Issue中提出或联系项目维护者。 