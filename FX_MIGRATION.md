# Uber-go/fx 依赖注入框架迁移指南

## 简介

本项目已引入 **uber-go/fx** 依赖注入框架，提供了两种启动方式：

1. **传统方式** (`cmd/server/main.go`) - 手动依赖注入
2. **Fx 方式** (`cmd/server/main_fx.go`) - 自动依赖注入和生命周期管理

## 为什么使用 Fx？

### 优势

- **自动依赖管理**: 无需手动创建和传递依赖，Fx 自动解析依赖图。
- **生命周期管理**: 统一管理组件的启动和关闭，确保资源正确释放。
- **模块化**: 轻松添加或移除功能模块（如 Redis, MongoDB）。
- **测试友好**: 可以轻松替换依赖进行单元测试。
- **并发安全**: Fx 确保依赖的初始化顺序正确。

### 对比

| 特性 | 传统方式 | Fx 方式 |
|------|---------|---------|
| 依赖注入 | 手动 | 自动 |
| 生命周期管理 | 手动编写启动/关闭逻辑 | 声明式钩子 |
| 代码量 | 较多 | 较少 |
| 模块化 | 需要手动管理 | 内置支持 |
| 学习曲线 | 低 | 中等 |

## 如何切换到 Fx 模式

### 步骤 1: 备份当前 main.go

```bash
mv cmd/server/main.go cmd/server/main_old.go
```

### 步骤 2: 启用 Fx 版本

```bash
mv cmd/server/main_fx.go cmd/server/main.go
```

### 步骤 3: 重新构建

```bash
go build -o bin/server cmd/server/main.go
```

### 步骤 4: 运行

```bash
./bin/server
```

## Fx 架构说明

### 依赖图

```
Config
  ├─> Database (PostgreSQL)
  ├─> Redis (可选)
  ├─> MongoDB (可选)
  ├─> Auth Service
  │     └─> JWT Generator
  ├─> User Repository
  │     └─> User Service
  │           └─> User Handler
  └─> HTTP Server
```

### 核心概念

#### 1. Provide（提供者）

使用 `fx.Provide` 注册依赖的构造函数。

```go
fx.Provide(
    func(cfg *config.Config) (*gorm.DB, error) {
        return db.NewPostgresDBFromDatabaseConfig(cfg.Database)
    },
)
```

#### 2. Invoke（调用者）

使用 `fx.Invoke` 触发依赖的初始化和使用。

```go
fx.Invoke(func(lc fx.Lifecycle, srv *http.Server) {
    // 启动逻辑
})
```

#### 3. Lifecycle（生命周期）

使用 `fx.Lifecycle` 管理组件的启动和关闭。

```go
lc.Append(fx.Hook{
    OnStart: func(ctx context.Context) error {
        // 启动逻辑
        return nil
    },
    OnStop: func(ctx context.Context) error {
        // 关闭逻辑
        return nil
    },
})
```

## 添加新模块

### 示例：添加 Email 服务

#### 1. 定义接口和实现

```go
// internal/email/service.go
package email

type Service interface {
    SendEmail(to, subject, body string) error
}

type service struct {
    smtpHost string
    smtpPort int
}

func NewService(smtpHost string, smtpPort int) Service {
    return &service{
        smtpHost: smtpHost,
        smtpPort: smtpPort,
    }
}

func (s *service) SendEmail(to, subject, body string) error {
    // 实现邮件发送逻辑
    return nil
}
```

#### 2. 在 main.go 中注册

```go
fx.Provide(
    func(cfg *config.Config) email.Service {
        return email.NewService(cfg.Email.Host, cfg.Email.Port)
    },
),
```

#### 3. 在需要的地方注入

```go
fx.Provide(
    func(userService user.Service, emailService email.Service) *user.Handler {
        return user.NewHandlerWithEmail(userService, emailService)
    },
),
```

## 可选模块管理

Fx 使得可选模块的管理变得简单。

### Redis 示例

```go
fx.Provide(
    func(cfg *config.Config) (*redis.Client, error) {
        if !cfg.Redis.Enabled {
            return nil, nil // 返回 nil 表示不启用
        }
        return redis.NewClient(cfg.Redis)
    },
),
```

在使用时：

```go
fx.Provide(
    func(repo user.Repository, redisClient *redis.Client) user.Service {
        if redisClient != nil {
            return user.NewServiceWithCache(repo, redisClient)
        }
        return user.NewService(repo)
    },
),
```

## 测试

Fx 使得测试变得更加容易。

### 单元测试

```go
func TestUserService(t *testing.T) {
    var svc user.Service

    app := fx.New(
        fx.Provide(
            func() user.Repository {
                return &mockUserRepository{} // 使用 mock
            },
        ),
        fx.Provide(user.NewService),
        fx.Populate(&svc), // 填充变量
    )

    app.Start(context.Background())
    defer app.Stop(context.Background())

    // 使用 svc 进行测试
}
```

## 最佳实践

1. **保持构造函数简单**: 构造函数应该只做依赖注入，不应该有复杂的初始化逻辑。
2. **使用接口**: 所有依赖都应该通过接口注入，便于测试和替换。
3. **避免循环依赖**: Fx 会检测并报告循环依赖。
4. **使用 fx.Module**: 对于大型应用，使用 `fx.Module` 组织相关的依赖。

## 故障排查

### 依赖未找到

**错误**: `missing type: *user.Handler`

**原因**: 没有为该类型提供构造函数。

**解决**: 添加 `fx.Provide` 注册构造函数。

### 循环依赖

**错误**: `cycle detected in dependency graph`

**原因**: A 依赖 B，B 又依赖 A。

**解决**: 重新设计依赖关系，引入接口打破循环。

## 回退到传统方式

如果遇到问题，可以随时回退：

```bash
mv cmd/server/main.go cmd/server/main_fx.go
mv cmd/server/main_old.go cmd/server/main.go
go build -o bin/server cmd/server/main.go
```

## 参考资源

- [Uber-go/fx 官方文档](https://uber-go.github.io/fx/)
- [Fx 最佳实践](https://github.com/uber-go/fx/blob/master/docs/best-practices.md)
