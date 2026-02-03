# Go REST API Starter

一个生产就绪的 Go REST API 微服务启动模板，专为微服务架构设计。

## 特性

✅ **清晰的架构** — Handler → Service → Repository (Go 行业标准)  
✅ **微服务网关集成** — 从网关获取用户信息（X-User-ID, X-User-Role）  
✅ **多数据库支持** — PostgreSQL、Redis、MongoDB  
✅ **数据库迁移** — 使用 golang-migrate 进行版本控制  
✅ **完善的测试** — 单元测试 + 集成测试  
✅ **结构化日志** — JSON 格式日志，带请求 ID  
✅ **标准化 API 响应** — 统一的响应格式  
✅ **生产级 Docker** — 多阶段构建，健康检查  
✅ **环境配置** — 支持开发/预发/生产环境  
✅ **优雅关闭** — 零停机部署  
✅ **热重载** — 开发环境 2 秒热重载

## 快速开始

### 前置要求

- [Docker](https://docs.docker.com/get-docker/) 和 [Docker Compose](https://docs.docker.com/compose/install/)
- [Git](https://git-scm.com/downloads)

### 一键启动

```bash
git clone https://github.com/yeegeek/go-rest-api-starter.git
cd go-rest-api-starter
make quick-start
```

**🎉 完成！** 您的 API 现在运行在：

- **API 基础 URL:** <http://localhost:8080/api/v1>
- **Swagger UI:** <http://localhost:8080/swagger/index.html>
- **健康检查:** <http://localhost:8080/health>

## 项目结构

```
.
├── cmd/                    # 应用程序入口
│   ├── server/            # API 服务器
│   ├── migrate/           # 数据库迁移工具
│   └── createadmin/       # 创建管理员工具
├── internal/              # 内部包
│   ├── auth/             # 认证相关（已适配网关模式）
│   ├── config/           # 配置管理
│   ├── db/               # 数据库连接
│   ├── errors/           # 错误处理
│   ├── health/           # 健康检查
│   ├── middleware/       # 中间件
│   ├── server/           # 路由设置
│   └── user/             # 用户模块示例
├── migrations/            # 数据库迁移文件
├── configs/              # 配置文件
├── scripts/              # 脚本文件
├── tests/                # 测试文件
└── Makefile              # Make 命令
```

## 微服务网关集成

本项目专为微服务架构设计，假设在 API 网关层已完成 JWT 认证。微服务从以下 HTTP 头获取用户信息：

- `X-User-ID`: 当前用户 ID
- `X-User-Role`: 用户角色（如：user, admin）

### 示例：Nginx 网关配置

```nginx
location /api/ {
    # JWT 验证（使用 auth_request 或 lua）
    auth_request /auth/verify;
    
    # 传递用户信息到后端微服务
    proxy_set_header X-User-ID $jwt_user_id;
    proxy_set_header X-User-Role $jwt_user_role;
    
    proxy_pass http://backend-service:8080;
}
```

### 示例：Kong 网关配置

```yaml
plugins:
  - name: jwt
  - name: request-transformer
    config:
      add:
        headers:
          - X-User-ID:$(jwt_claims.sub)
          - X-User-Role:$(jwt_claims.role)
```

## 数据库支持

### PostgreSQL

默认主数据库，用于关系型数据存储。

```yaml
database:
  host: "db"
  port: 5432
  user: "postgres"
  password: "your-password"
  name: "go_rest_api_starter"
```

### Redis

用于缓存和会话存储。

```yaml
redis:
  host: "redis"
  port: 6379
  password: ""
  db: 0
```

### MongoDB

用于文档型数据存储。

```yaml
mongodb:
  uri: "mongodb://mongodb:27017"
  database: "go_rest_api_starter"
```

## 常用命令

```bash
# 开发
make dev                    # 启动开发环境（热重载）
make build                  # 编译应用
make test                   # 运行测试
make test-coverage          # 测试覆盖率

# 数据库迁移
make migrate-up             # 执行所有待执行的迁移
make migrate-down           # 回滚最后一次迁移
make migrate-create NAME=xxx # 创建新的迁移文件
make migrate-status         # 查看迁移状态

# Docker
make docker-build           # 构建 Docker 镜像
make docker-up              # 启动所有服务
make docker-down            # 停止所有服务
make docker-logs            # 查看日志

# 清理
make clean                  # 清理构建文件
```

## 环境配置

项目支持多环境配置：

- `config.yaml` - 基础配置
- `config.development.yaml` - 开发环境
- `config.staging.yaml` - 预发环境
- `config.production.yaml` - 生产环境

通过环境变量 `APP_ENVIRONMENT` 切换环境：

```bash
export APP_ENVIRONMENT=production
```

## API 文档

启动服务后访问 Swagger UI：

<http://localhost:8080/swagger/index.html>

或导入 Postman 集合：

```bash
api/postman_collection.json
```

## 健康检查

- `/health` - 整体健康状态
- `/health/live` - 存活探针（Kubernetes liveness）
- `/health/ready` - 就绪探针（Kubernetes readiness）

## 开发指南

### 添加新模块

1. 在 `internal/` 下创建新包
2. 实现 Handler、Service、Repository 三层
3. 在 `internal/server/router.go` 注册路由
4. 创建数据库迁移文件（如需要）
5. 编写单元测试和集成测试

### 代码规范

- 遵循 Go 官方代码规范
- 使用 `gofmt` 格式化代码
- 使用 `golangci-lint` 进行代码检查
- 保持测试覆盖率 > 80%

## 部署

### Docker 部署

```bash
docker-compose -f docker-compose.prod.yml up -d
```

### Kubernetes 部署

```bash
# 构建镜像
docker build -t go-rest-api-starter:latest .

# 推送到镜像仓库
docker push your-registry/go-rest-api-starter:latest

# 部署到 K8s
kubectl apply -f k8s/
```

## 安全建议

1. **生产环境必须配置**：
   - 数据库密码
   - Redis 密码（如使用）
   - MongoDB 认证（如使用）

2. **网关层安全**：
   - 确保网关正确验证 JWT
   - 防止 X-User-ID 和 X-User-Role 头被客户端伪造
   - 使用 HTTPS

3. **CORS 配置**：
   - 生产环境限制允许的域名
   - 不要使用 `AllowAllOrigins`

4. **Rate Limiting**：
   - 根据实际负载调整限流参数
   - 考虑使用分布式限流（Redis）

## 性能优化

1. **数据库连接池**：根据负载调整 `MaxOpenConns` 和 `MaxIdleConns`
2. **Redis 缓存**：缓存频繁查询的数据
3. **MongoDB 索引**：为常用查询字段创建索引
4. **日志级别**：生产环境使用 `info` 或 `warn` 级别

## 监控和日志

### 结构化日志

所有日志以 JSON 格式输出，包含：

- `timestamp`: 时间戳
- `level`: 日志级别
- `message`: 日志消息
- `request_id`: 请求 ID（用于追踪）
- `user_id`: 用户 ID（如有）

### 指标监控

建议集成：

- Prometheus - 指标收集
- Grafana - 可视化
- Jaeger - 分布式追踪

## 许可证

MIT License

## 贡献

欢迎提交 Issue 和 Pull Request！
