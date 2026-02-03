package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"go.uber.org/fx"
	"gorm.io/gorm"

	_ "github.com/yeegeek/go-rest-api-starter/api/docs"
	"github.com/yeegeek/go-rest-api-starter/internal/auth"
	"github.com/yeegeek/go-rest-api-starter/internal/config"
	"github.com/yeegeek/go-rest-api-starter/internal/db"
	"github.com/yeegeek/go-rest-api-starter/internal/migrate"
	"github.com/yeegeek/go-rest-api-starter/internal/mongodb"
	"github.com/yeegeek/go-rest-api-starter/internal/redis"
	"github.com/yeegeek/go-rest-api-starter/internal/server"
	"github.com/yeegeek/go-rest-api-starter/internal/user"
)

// 使用 uber-go/fx 进行依赖注入的新版本 main
// 要使用此版本，请将 main.go 重命名为 main_old.go，并将此文件重命名为 main.go

func mainWithFx() {
	app := fx.New(
		// 提供配置
		fx.Provide(
			func() (*config.Config, error) {
				cfg, err := config.LoadConfig("")
				if err != nil {
					return nil, err
				}
				if err := cfg.Validate(); err != nil {
					return nil, err
				}
				return cfg, nil
			},
		),

		// 提供日志器
		fx.Provide(
			func() *slog.Logger {
				return slog.Default()
			},
		),

		// 提供数据库连接
		fx.Provide(
			func(cfg *config.Config) (*gorm.DB, error) {
				return db.NewPostgresDBFromDatabaseConfig(cfg.Database)
			},
		),

		// 提供 Redis 客户端（可选）
		fx.Provide(
			func(cfg *config.Config) (*redis.Client, error) {
				if !cfg.Redis.Enabled {
					return nil, nil
				}
				return redis.NewClient(cfg.Redis)
			},
		),

		// 提供 MongoDB 客户端（可选）
		fx.Provide(
			func(cfg *config.Config) (*mongodb.Client, error) {
				if !cfg.MongoDB.Enabled {
					return nil, nil
				}
				return mongodb.NewClient(context.Background(), cfg.MongoDB)
			},
		),

		// 提供 Auth Service
		fx.Provide(
			func(cfg *config.Config, db *gorm.DB) auth.Service {
				return auth.NewServiceWithRepo(&cfg.JWT, db)
			},
		),

		// 提供 JWT Generator
		fx.Provide(
			func(cfg *config.Config, db *gorm.DB) auth.JWTGenerator {
				return auth.NewJWTGenerator(&cfg.JWT, db)
			},
		),

		// 提供 User 模块
		fx.Provide(
			func(db *gorm.DB) user.Repository {
				return user.NewRepository(db)
			},
		),
		fx.Provide(
			func(repo user.Repository) user.Service {
				return user.NewService(repo)
			},
		),
		fx.Provide(
			func(userService user.Service, authService auth.Service) *user.Handler {
				return user.NewHandler(userService, authService)
			},
		),

		// 提供 HTTP 服务器
		fx.Provide(
			func(
				userHandler *user.Handler,
				authService auth.Service,
				cfg *config.Config,
				db *gorm.DB,
			) *http.Server {
				router := server.SetupRouter(userHandler, authService, cfg, db)

				port := cfg.Server.Port
				if port == "" {
					port = "8080"
				}

				maxHeaderBytes := cfg.Server.MaxHeaderBytes
				if maxHeaderBytes == 0 {
					maxHeaderBytes = 1 << 20
				}

				return &http.Server{
					Addr:           fmt.Sprintf(":%s", port),
					Handler:        router,
					ReadTimeout:    time.Duration(cfg.Server.ReadTimeout) * time.Second,
					WriteTimeout:   time.Duration(cfg.Server.WriteTimeout) * time.Second,
					IdleTimeout:    time.Duration(cfg.Server.IdleTimeout) * time.Second,
					MaxHeaderBytes: maxHeaderBytes,
				}
			},
		),

		// 启动和停止钩子
		fx.Invoke(func(lc fx.Lifecycle, srv *http.Server, cfg *config.Config, db *gorm.DB, logger *slog.Logger) {
			lc.Append(fx.Hook{
				OnStart: func(ctx context.Context) error {
					logger.Info("Starting Go REST API Starter...")
					cfg.LogSafeConfig(logger)

					// 检查迁移状态
					if err := checkMigrationStatus(db, &cfg.Migrations); err != nil {
						logger.Warn("Migration check", "status", "⚠️", "error", err)
					} else {
						logger.Info("Migration check", "status", "✓")
					}

					// 启动 HTTP 服务器
					go func() {
						port := cfg.Server.Port
						if port == "" {
							port = "8080"
						}
						logger.Info("Server starting", "address", srv.Addr)
						logger.Info("Swagger UI available", "url", fmt.Sprintf("http://localhost:%s/swagger/index.html", port))
						logger.Info("Health check available", "url", fmt.Sprintf("http://localhost:%s/health", port))

						if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
							logger.Error("Server error", "error", err)
						}
					}()

					return nil
				},
				OnStop: func(ctx context.Context) error {
					logger.Info("Shutting down server gracefully...")

					// 关闭数据库连接
					sqlDB, err := db.DB()
					if err == nil {
						logger.Info("Closing database connections...")
						if err := sqlDB.Close(); err != nil {
							logger.Error("Error closing database", "error", err)
						}
					}

					// 关闭 HTTP 服务器
					if err := srv.Shutdown(ctx); err != nil {
						logger.Error("Server forced to shutdown", "error", err)
						return err
					}

					logger.Info("Server exited gracefully")
					return nil
				},
			})
		}),
	)

	app.Run()
}

func checkMigrationStatus(database *gorm.DB, cfg *config.MigrationsConfig) error {
	sqlDB, err := database.DB()
	if err != nil {
		return fmt.Errorf("failed to get sql.DB: %w", err)
	}

	migrator, err := migrate.New(sqlDB, migrate.Config{
		MigrationsDir: cfg.Directory,
		Timeout:       time.Duration(cfg.Timeout) * time.Second,
		LockTimeout:   time.Duration(cfg.LockTimeout) * time.Second,
	})
	if err != nil {
		return fmt.Errorf("failed to create migrator: %w", err)
	}

	version, dirty, err := migrator.Version()
	if err != nil {
		return fmt.Errorf("failed to get migration version: %w", err)
	}

	if dirty {
		return fmt.Errorf("database in dirty state at version %d", version)
	}

	slog.Info("Database schema", "version", version)
	return nil
}
