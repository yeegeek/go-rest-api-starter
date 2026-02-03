# Go REST API Starter 用户手册

## 1. 简介

本文档是 **Go REST API Starter** 的详细用户手册。这是一个生产就绪的 Go REST API 微服务启动模板，经过深度改造，专为微服务架构设计。它移除了原有的本地 JWT 认证，改为依赖 API 网关传递用户信息，并集成了 Redis 和 MongoDB 支持。

### 1.1. 核心设计理念

- **微服务优先**: 假设服务运行在 API 网关之后，认证和部分流量控制由网关处理。
- **关注点分离**: 严格遵循 `Handler` -> `Service` -> `Repository` 的分层架构。
- **可扩展性**: 轻松添加新的数据源（如 Redis, MongoDB）和业务模块。
- **生产就绪**: 提供结构化日志、健康检查、多环境配置和 Docker 支持。

### 1.2. 主要特性

| 特性 | 描述 |
| :--- | :--- |
| **清晰架构** | Handler → Service → Repository (Go 行业标准)。 |
| **网关集成** | 从 `X-User-ID` 和 `X-User-Role` 头获取用户信息。 |
| **多数据库** | 内置 PostgreSQL, Redis, MongoDB 支持，可按需启用。 |
| **数据库迁移** | 使用 `golang-migrate` 进行 SQL 数据库的版本控制。 |
| **结构化日志** | JSON 格式日志，包含请求 ID，便于追踪。 |
| **标准化响应** | 所有 API 响应都使用统一的 `{success, data, error, meta}` 格式。 |
| **Docker 支持** | 提供开发和生产环境的 Docker 和 Docker Compose 配置。 |
| **热重载** | 开发环境下，代码变更可在 2 秒内自动重新加载。 |

---

## 2. 项目结构

项目遵循标准的 Go 项目布局，核心业务逻辑位于 `internal` 目录中。

```
.
├── cmd/                    # 应用程序入口
│   ├── server/            # API 服务器主程序 (main.go)
│   ├── migrate/           # 数据库迁移工具
│   └── createadmin/       # 创建管理员工具
├── configs/              # 配置文件 (config.yaml, config.production.yaml等)
├── internal/              # 内部业务逻辑 (不对外暴露)
│   ├── config/           # 配置加载和验证
│   ├── contextutil/      # 从 Gin 上下文获取用户信息的辅助函数
│   ├── db/               # PostgreSQL (GORM) 连接
│   ├── errors/           # 自定义错误处理和标准化响应
│   ├── health/           # 健康检查端点 (/health, /live, /ready)
│   ├── middleware/       # Gin 中间件 (日志, 错误处理, 网关认证)
│   ├── mongodb/          # MongoDB 连接和客户端封装
│   ├── redis/            # Redis 连接和客户端封装
│   ├── server/           # Gin 路由设置
│   └── user/             # 用户模块示例 (Handler, Service, Repository)
├── migrations/            # SQL 数据库迁移文件 (*.up.sql, *.down.sql)
├── scripts/              # Shell 脚本 (如: 快速启动脚本)
├── tests/                # 集成测试和端到端测试
├── docker-compose.yml    # 开发环境 Docker Compose
├── docker-compose.prod.yml # 生产环境 Docker Compose
├── Dockerfile            # 多阶段 Dockerfile
├── go.mod                # Go 模块依赖
└── Makefile              # 常用命令自动化
```

---

## 3. 快速开始

### 3.1. 前置要求

- **Git**: 用于克隆代码库。
- **Docker**: 用于运行容器化环境。
- **Docker Compose**: 用于编排多容器应用。
- **Make**: 用于执行 `Makefile` 中的命令。

### 3.2. 一键启动

此命令将使用 Docker Compose 启动开发环境，包括 API 服务、PostgreSQL、Redis 和 MongoDB。

```bash
# 1. 克隆仓库
git clone https://github.com/yeegeek/go-rest-api-starter.git

# 2. 进入目录
cd go-rest-api-starter

# 3. 启动服务
make quick-start
```

启动后，服务将在以下地址可用：

- **API 基础 URL**: `http://localhost:8080/api/v1`
- **Swagger UI**: `http://localhost:8080/swagger/index.html`
- **健康检查**: `http://localhost:8080/health`

---

## 4. 配置管理

项目使用 [Viper](https://github.com/spf13/viper) 进行配置管理，支持多文件和环境变量覆盖。

### 4.1. 配置文件

- `configs/config.yaml`: 基础配置，适用于所有环境。
- `configs/config.development.yaml`: 开发环境配置，会覆盖基础配置。
- `configs/config.staging.yaml`: 预发环境配置。
- `configs/config.production.yaml`: 生产环境配置。

通过设置 `APP_ENVIRONMENT` 环境变量来选择加载哪个环境的配置。

### 4.2. 环境变量

所有在 `.yaml` 文件中的配置都可以通过环境变量覆盖。命名规则为：将 `.` 替换为 `_` 并大写。

**示例**: 要覆盖数据库主机，设置 `DATABASE_HOST=my-db-host`。

### 4.3. 数据库配置

你可以在 `config.*.yaml` 文件中启用或禁用 Redis 和 MongoDB。

```yaml
# config.yaml

redis:
  enabled: true  # 设置为 true 来启用 Redis
  host: "redis"
  port: 6379
  # ...

mongodb:
  enabled: true  # 设置为 true 来启用 MongoDB
  uri: "mongodb://mongodb:27017"
  database: "go_rest_api_starter"
```

---

## 5. 核心功能改造

### 5.1. 网关认证模式

本项目已移除所有本地 JWT 生成和验证逻辑。它依赖于上游的 API 网关（如 Nginx, Kong, Traefik）来验证用户身份，并通过 HTTP 头将用户信息传递给本服务。

- **`X-User-ID`**: 用户的唯一标识符。
- **`X-User-Role`**: 用户的角色（例如 `user`, `admin`）。

`internal/middleware/gateway_auth.go` 中间件负责从这些头中读取信息，并将其存入 Gin 的上下文中。所有受保护的路由都必须使用此中间件。

**示例路由定义**:
```go
// internal/server/router.go

// 用户端点 - 需要网关认证
usersGroup := v1.Group("/users")
usersGroup.Use(middleware.GatewayAuthMiddleware())
{
    usersGroup.GET("/me", userHandler.GetMe)
}

// 管理员端点 - 需要网关认证和管理员角色
adminGroup := v1.Group("/admin")
adminGroup.Use(middleware.GatewayAuthMiddleware(), middleware.RequireAdminRole())
{
    adminGroup.GET("/users", userHandler.ListUsers)
}
```

### 5.2. Redis 支持

通过 `internal/redis/redis.go` 提供了一个 Redis 客户端封装。你可以在需要缓存的 Service 层注入并使用它。

**使用示例**:
```go
// 假设在 Service 中注入了 redisClient *redis.Client

func (s *myService) GetData(ctx context.Context, key string) (string, error) {
    // 1. 尝试从缓存获取
    cachedValue, err := s.redisClient.Get(ctx, key)
    if err != nil && err != redis.Nil {
        return "", err
    }
    if cachedValue != "" {
        return cachedValue, nil
    }

    // 2. 从数据库获取
    dbValue, err := s.repo.GetValue(ctx, key)
    if err != nil {
        return "", err
    }

    // 3. 存入缓存
    _ = s.redisClient.Set(ctx, key, dbValue, 10*time.Minute)

    return dbValue, nil
}
```

### 5.3. MongoDB 支持

通过 `internal/mongodb/mongodb.go` 提供了一个 MongoDB 客户端封装。你可以用它来处理非结构化或文档型数据。

**使用示例**:
```go
// 假设在 Service 中注入了 mongoClient *mongodb.Client

import "go.mongodb.org/mongo-driver/bson"

type LogEntry struct {
    Level     string    `bson:"level"`
    Message   string    `bson:"message"`
    Timestamp time.Time `bson:"timestamp"`
}

func (s *myService) SaveLog(ctx context.Context, entry LogEntry) error {
    _, err := s.mongoClient.InsertOne(ctx, "audit_logs", entry)
    return err
}

func (s *myService) GetLogs(ctx context.Context, level string) ([]LogEntry, error) {
    cursor, err := s.mongoClient.Find(ctx, "audit_logs", bson.M{"level": level})
    if err != nil {
        return nil, err
    }
    defer cursor.Close(ctx)

    var results []LogEntry
    if err = cursor.All(ctx, &results); err != nil {
        return nil, err
    }
    return results, nil
}
```

---

## 6. 开发指南

### 6.1. 添加新模块

以添加一个 `product` 模块为例：

1.  **创建 Model** (`internal/product/model.go`):
    定义 `Product` 结构体和 GORM 模型。

2.  **创建 Repository** (`internal/product/repository.go`):
    实现数据库的增删改查操作，定义 `Repository` 接口。

3.  **创建 Service** (`internal/product/service.go`):
    实现业务逻辑，如价格计算、库存检查等，定义 `Service` 接口。

4.  **创建 Handler** (`internal/product/handler.go`):
    处理 HTTP 请求，验证输入，调用 Service，并返回 JSON 响应。

5.  **注册路由** (`internal/server/router.go`):
    将新的 Handler 和路由添加到 Gin 路由器中。

6.  **数据库迁移**:
    ```bash
    make migrate-create NAME=create_products_table
    ```
    编辑生成的 `.sql` 文件来创建 `products` 表。

7.  **编写测试**: 为 Repository, Service, Handler 编写单元测试和集成测试。

### 6.2. 常用命令

`Makefile` 提供了丰富的命令来简化开发流程。

- `make dev`: 启动开发环境（带热重载）。
- `make test`: 运行所有测试。
- `make test-coverage`: 计算测试覆盖率并生成报告。
- `make lint`: 运行 `golangci-lint` 进行代码检查。
- `make docker-build`: 构建生产环境的 Docker 镜像。
- `make docker-up-prod`: 启动生产环境的 Docker Compose 服务。

---

## 7. 部署

### 7.1. Docker Compose 部署

对于简单的生产部署，可以使用 `docker-compose.prod.yml`。

```bash
# 确保在 .env 文件或环境变量中设置了生产密码
export DATABASE_PASSWORD=your_strong_password

# 启动服务
docker-compose -f docker-compose.prod.yml up -d
```

### 7.2. Kubernetes 部署

1.  **构建并推送镜像**:
    ```bash
    docker build -t your-registry/go-rest-api-starter:latest .
    docker push your-registry/go-rest-api-starter:latest
    ```

2.  **配置 Kubernetes**:
    - 创建 `Secret` 来存储数据库密码等敏感信息。
    - 创建 `ConfigMap` 来存储非敏感配置。
    - 创建 `Deployment` 来部署应用。
    - 创建 `Service` 来暴露应用。
    - 创建 `Ingress` 来管理外部访问（并配置网关逻辑）。

---

## 8. 安全建议

- **网关安全**: 确保 API 网关能有效防止 `X-User-ID` 和 `X-User-Role` 头被客户端直接伪造。
- **生产密码**: 绝不在代码或配置文件中硬编码生产密码，始终使用环境变量或 Secrets Management 工具。
- **CORS 策略**: 在生产环境中，将 `corsConfig.AllowAllOrigins` 设置为 `false`，并明确指定允许的前端域名。
- **输入验证**: 尽管本项目有基础的验证，但对所有来自外部的输入（参数、请求体）都应进行严格的验证、清理和转义。
