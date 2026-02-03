package server

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/gorm"

	"github.com/yeegeek/go-rest-api-starter/internal/auth"
	"github.com/yeegeek/go-rest-api-starter/internal/config"
	"github.com/yeegeek/go-rest-api-starter/internal/errors"
	"github.com/yeegeek/go-rest-api-starter/internal/health"
	"github.com/yeegeek/go-rest-api-starter/internal/middleware"
	"github.com/yeegeek/go-rest-api-starter/internal/user"
)

// SetupRouter creates and configures the Gin router
func SetupRouter(userHandler *user.Handler, authService auth.Service, cfg *config.Config, db *gorm.DB) *gin.Engine {
	router := gin.New()

	if cfg.App.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	skipPaths := config.GetSkipPaths(cfg.App.Environment)
	loggerConfig := middleware.NewLoggerConfig(
		cfg.Logging.GetLogLevel(),
		skipPaths,
	)
	// 日志、错误处理和恢复中间件
	router.Use(middleware.Logger(loggerConfig))
	router.Use(errors.ErrorHandler())
	router.Use(gin.Recovery())

	// 输入验证中间件 - 防止 SQL 注入和 XSS 攻击
	router.Use(middleware.InputValidationMiddleware())

	// CORS 和 Rate Limiting 由 API 网关处理

	var checkers []health.Checker
	if cfg.Health.DatabaseCheckEnabled {
		dbChecker := health.NewDatabaseChecker(db)
		checkers = append(checkers, dbChecker)
	}
	healthService := health.NewService(checkers, cfg.App.Version, cfg.App.Environment)
	healthHandler := health.NewHandler(healthService)

	router.GET("/health", healthHandler.Health)
	router.GET("/health/live", healthHandler.Live)
	router.GET("/health/ready", healthHandler.Ready)

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	v1 := router.Group("/api/v1")
	{
		// 公开端点（无需认证）
		publicGroup := v1.Group("/public")
		{
			publicGroup.POST("/register", userHandler.Register)
		}

		// 用户端点 - 需要网关认证
		usersGroup := v1.Group("/users")
		usersGroup.Use(middleware.GatewayAuthMiddleware())
		{
			usersGroup.GET("/me", userHandler.GetMe)
			usersGroup.GET("/:id", userHandler.GetUser)
			usersGroup.PUT("/:id", userHandler.UpdateUser)
			usersGroup.DELETE("/:id", userHandler.DeleteUser)
		}

		// 管理员端点 - 需要网关认证和管理员角色
		adminGroup := v1.Group("/admin")
		adminGroup.Use(middleware.GatewayAuthMiddleware(), middleware.RequireAdminRole())
		{
			// 用户管理端点
			adminGroup.GET("/users", userHandler.ListUsers)
			adminGroup.GET("/users/:id", userHandler.GetUser)
			adminGroup.PUT("/users/:id", userHandler.UpdateUser)
			adminGroup.DELETE("/users/:id", userHandler.DeleteUser)
		}
	}

	return router
}
