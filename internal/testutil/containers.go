package testutil

import (
	"context"
	"fmt"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/modules/redis"
	"github.com/testcontainers/testcontainers-go/wait"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	redisClient "github.com/redis/go-redis/v9"
)

// PostgresContainer 封装 PostgreSQL 测试容器
type PostgresContainer struct {
	Container *postgres.PostgresContainer
	DB        *gorm.DB
	DSN       string
}

// NewPostgresContainer 创建新的 PostgreSQL 测试容器
func NewPostgresContainer(ctx context.Context) (*PostgresContainer, error) {
	pgContainer, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to start postgres container: %w", err)
	}

	// 获取连接字符串
	dsn, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		return nil, fmt.Errorf("failed to get connection string: %w", err)
	}

	// 连接数据库
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to test database: %w", err)
	}

	return &PostgresContainer{
		Container: pgContainer,
		DB:        db,
		DSN:       dsn,
	}, nil
}

// Close 关闭并清理容器
func (pc *PostgresContainer) Close(ctx context.Context) error {
	if pc.Container != nil {
		return pc.Container.Terminate(ctx)
	}
	return nil
}

// RedisContainer 封装 Redis 测试容器
type RedisContainer struct {
	Container testcontainers.Container
	Client    *redisClient.Client
	URI       string
}

// NewRedisContainer 创建新的 Redis 测试容器
func NewRedisContainer(ctx context.Context) (*RedisContainer, error) {
	req := testcontainers.ContainerRequest{
		Image:        "redis:7-alpine",
		ExposedPorts: []string{"6379/tcp"},
		WaitingFor: wait.ForLog("Ready to accept connections").
			WithStartupTimeout(30 * time.Second),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to start redis container: %w", err)
	}

	// 获取主机和端口
	host, err := container.Host(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get container host: %w", err)
	}

	mappedPort, err := container.MappedPort(ctx, "6379")
	if err != nil {
		return nil, fmt.Errorf("failed to get mapped port: %w", err)
	}

	// 创建 Redis 客户端
	uri := fmt.Sprintf("redis://%s:%s", host, mappedPort.Port())
	client := redisClient.NewClient(&redisClient.Options{
		Addr: fmt.Sprintf("%s:%s", host, mappedPort.Port()),
	})

	// 测试连接
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to ping redis: %w", err)
	}

	return &RedisContainer{
		Container: container,
		Client:    client,
		URI:       uri,
	}, nil
}

// Close 关闭并清理容器
func (rc *RedisContainer) Close(ctx context.Context) error {
	if rc.Client != nil {
		_ = rc.Client.Close()
	}
	if rc.Container != nil {
		return rc.Container.Terminate(ctx)
	}
	return nil
}
