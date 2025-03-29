package main

import (
	"codefolio/internal/config"
	"codefolio/internal/domain"
	"codefolio/internal/handler"
	"codefolio/internal/repository"
	"codefolio/internal/service"
	"codefolio/internal/util"
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// 初始化日志
	logger := util.InitLogger()
	defer logger.Sync()

	// 加载配置
	cfg := config.LoadConfig()

	// 设置Gin模式
	if cfg.Server.Mode == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 配置文件上传参数
	util.SetUploadConfig(
		cfg.Upload.StoragePath,
		cfg.Upload.MaxFileSize,
		cfg.Upload.AllowedTypes,
	)

	// 创建路由
	r := gin.New()

	// 使用中间件
	r.Use(handler.LoggerMiddleware())
	r.Use(handler.Recovery())
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Content-Length", "Accept-Encoding", "X-CSRF-Token", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// 健康检查接口
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":      "online",
			"version":     cfg.Server.Version,
			"environment": cfg.Server.Mode,
		})
	})

	// 连接数据库
	logger.Info("正在连接数据库...", zap.String("dsn", cfg.GetDSN()))
	db, err := gorm.Open(postgres.Open(cfg.GetDSN()), &gorm.Config{})
	if err != nil {
		logger.Fatal("数据库连接失败", zap.Error(err))
	}

	// 自动迁移数据模型
	logger.Info("正在进行数据库迁移...")
	err = db.AutoMigrate(
		&domain.User{},
		&domain.Resume{},
		&domain.Tag{},
		&domain.Offer{},
		&domain.ResumeTag{},
	)
	if err != nil {
		logger.Fatal("数据库迁移失败", zap.Error(err))
	}

	// 初始化存储目录
	if err := os.MkdirAll(util.UploadDir, 0755); err != nil {
		logger.Fatal("创建上传目录失败", zap.Error(err))
	}

	// 创建仓库
	userRepo := repository.NewUserRepository(db)
	resumeRepo := repository.NewResumeRepository(db)

	// 创建服务
	userService := service.NewUserService(userRepo, cfg.JWT.Secret, cfg.JWT.ExpireHours)
	resumeService := service.NewResumeService(
		resumeRepo,
		userRepo,
		cfg.Upload.AnonymousView,
		cfg.Upload.UserView,
	)

	// 创建处理器
	userHandler := handler.NewUserHandler(userService)
	faqHandler := handler.NewFAQHandler()
	resumeHandler := handler.NewResumeHandler(resumeService)

	// 创建API分组
	api := r.Group("/api/v1")

	// 静态文件服务
	api.GET("/files/*path", resumeHandler.ServeResumeFile)

	// 用户相关路由
	api.POST("/register", userHandler.Register)
	api.POST("/login", userHandler.Login)
	api.GET("/me", handler.AuthMiddleware(cfg.JWT.Secret), userHandler.GetMe)

	// FAQ相关路由
	api.GET("/faqs", faqHandler.GetFAQs)

	// 简历相关路由
	resumeGroup := api.Group("/resumes")
	{
		// 公开路由
		resumeGroup.GET("", resumeHandler.GetResumes)
		resumeGroup.GET("/:id", resumeHandler.GetResume)
		resumeGroup.GET("/:id/download", resumeHandler.DownloadResume)
		resumeGroup.GET("/tags", resumeHandler.GetTags)

		// 需要认证的路由
		auth := resumeGroup.Use(handler.AuthMiddleware(cfg.JWT.Secret))
		{
			auth.POST("", resumeHandler.UploadResume)
			auth.PUT("/:id", resumeHandler.UpdateResume)
			auth.PUT("/:id/file", resumeHandler.UpdateResumeFile)
			auth.DELETE("/:id", resumeHandler.DeleteResume)
			auth.GET("/user/list", resumeHandler.GetUserResumes)
		}
	}

	// 启动服务器
	port := fmt.Sprintf(":%d", cfg.Server.Port)
	logger.Info("服务器已启动", zap.String("地址", port))

	srv := &http.Server{
		Addr:    port,
		Handler: r,
	}

	// 优雅关闭
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("启动服务器失败", zap.Error(err))
		}
	}()

	// 等待中断信号优雅关闭服务器
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("关闭服务器...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("服务器关闭出错", zap.Error(err))
	}

	logger.Info("服务器已关闭")
}
