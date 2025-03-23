package main

import (
	"codefolio/internal/common"
	"codefolio/internal/config"
	"codefolio/internal/domain"
	"codefolio/internal/handler"
	"codefolio/internal/repository"
	"codefolio/internal/service"
	"codefolio/internal/util"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// 加载配置
	cfg := config.LoadConfig()

	// 初始化日志
	logger := util.InitLogger(cfg.Server.Environment)
	defer logger.Sync() // 确保日志缓冲区被刷新

	// 设置运行模式
	if cfg.Server.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 连接数据库
	logger.Info("Connecting to database...",
		zap.String("host", cfg.Database.Host),
		zap.String("name", cfg.Database.Name))

	db, err := gorm.Open(postgres.Open(cfg.GetDSN()), &gorm.Config{})
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}

	// 自动迁移数据库
	logger.Info("Running database migrations...")
	if err := db.AutoMigrate(&domain.User{}); err != nil {
		logger.Fatal("Failed to migrate database", zap.Error(err))
	}

	// 初始化依赖
	userRepo := repository.NewUserRepository(db)
	userService := service.NewUserService(userRepo, cfg.JWT.Secret)
	userHandler := handler.NewUserHandler(userService)

	// 初始化FAQ处理器
	faqHandler := handler.NewFAQHandler()

	// 设置路由
	r := gin.New() // 不使用默认中间件，手动添加

	// 添加中间件
	r.Use(util.LoggerMiddleware())
	r.Use(gin.Recovery())
	r.Use(handler.CORSMiddleware())

	// 健康检查路由
	r.GET("/health", func(c *gin.Context) {
		common.ResponseWithData(c, gin.H{
			"status":      "ok",
			"version":     cfg.Server.Version,
			"environment": cfg.Server.Environment,
		})
	})

	// API路由组
	api := r.Group("/api/v1")

	// 公开路由
	{
		api.POST("/auth/register", userHandler.Register)
		api.POST("/auth/login", userHandler.Login)

		// FAQ路由 - 只提供获取所有FAQ的接口
		api.GET("/faqs", faqHandler.GetAllFAQs)
	}

	// 需要认证的路由
	authorized := api.Group("")
	authorized.Use(handler.AuthMiddleware(cfg.JWT.Secret))
	{
		authorized.GET("/auth/me", userHandler.GetMe)
	}

	// 启动服务器
	logger.Info("Starting server",
		zap.String("port", cfg.Server.Port),
		zap.String("environment", cfg.Server.Environment),
		zap.String("version", cfg.Server.Version))

	if err := r.Run(":" + cfg.Server.Port); err != nil {
		logger.Fatal("Failed to start server", zap.Error(err))
	}
}
