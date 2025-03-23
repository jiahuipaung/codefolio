package main

import (
	"codefolio/internal/domain"
	"codefolio/internal/handler"
	"codefolio/internal/repository"
	"codefolio/internal/service"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// 加载环境变量
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	// 连接数据库
	dsn := "host=" + os.Getenv("DB_HOST") + " user=" + os.Getenv("DB_USER") + " password=" + os.Getenv("DB_PASSWORD") + " dbname=" + os.Getenv("DB_NAME") + " port=" + os.Getenv("DB_PORT") + " sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// 自动迁移数据库
	if err := db.AutoMigrate(&domain.User{}); err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	// 初始化依赖
	userRepo := repository.NewUserRepository(db)
	userService := service.NewUserService(userRepo, os.Getenv("JWT_SECRET"))
	userHandler := handler.NewUserHandler(userService)

	// 设置路由
	r := gin.Default()

	// 公开路由
	public := r.Group("/api/v1")
	{
		public.POST("/auth/register", userHandler.Register)
		public.POST("/auth/login", userHandler.Login)
	}

	// 需要认证的路由
	authorized := r.Group("/api/v1")
	authorized.Use(handler.AuthMiddleware(os.Getenv("JWT_SECRET")))
	{
		authorized.GET("/auth/me", userHandler.GetMe)
	}

	// 启动服务器
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	if err := r.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
