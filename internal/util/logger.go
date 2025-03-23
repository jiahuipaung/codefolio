package util

import (
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var logger *zap.Logger

// InitLogger 初始化日志记录器
func InitLogger() *zap.Logger {
	// 创建日志目录
	logDir := "logs"
	if err := os.MkdirAll(logDir, 0755); err != nil {
		panic("无法创建日志目录: " + err.Error())
	}

	// 获取环境变量
	env := os.Getenv("SERVER_MODE")
	if env == "" {
		env = "development" // 默认为开发环境
	}

	// 设置日志配置
	var config zap.Config
	if env == "production" {
		// 生产环境: JSON格式，写入文件
		config = zap.NewProductionConfig()
		config.OutputPaths = []string{
			"stdout",
			filepath.Join(logDir, "app.log"),
		}
		config.ErrorOutputPaths = []string{
			"stderr",
			filepath.Join(logDir, "error.log"),
		}
	} else {
		// 开发环境: 彩色输出到控制台
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	// 构建日志记录器
	var err error
	logger, err = config.Build()
	if err != nil {
		panic("初始化日志记录器失败: " + err.Error())
	}

	logger.Info("日志系统已初始化",
		zap.String("environment", env),
		zap.String("logDir", logDir),
	)

	return logger
}

// GetLogger 获取全局日志记录器
func GetLogger() *zap.Logger {
	if logger == nil {
		// 如果日志记录器未初始化，使用默认配置
		logger, _ = zap.NewProduction()
	}
	return logger
}

// LoggerMiddleware 日志中间件
func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 开始时间
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		// 处理请求
		c.Next()

		// 结束时间
		end := time.Now()
		latency := end.Sub(start)

		// 获取状态码和错误信息
		status := c.Writer.Status()
		clientIP := c.ClientIP()
		method := c.Request.Method

		// 构建日志字段
		fields := []zapcore.Field{
			zap.Int("status", status),
			zap.String("method", method),
			zap.String("path", path),
			zap.String("query", query),
			zap.String("ip", clientIP),
			zap.Duration("latency", latency),
			zap.String("user-agent", c.Request.UserAgent()),
		}

		// 根据状态码记录不同级别的日志
		switch {
		case status >= http.StatusInternalServerError:
			GetLogger().Error("服务器错误", fields...)
		case status >= http.StatusBadRequest:
			GetLogger().Warn("客户端错误", fields...)
		default:
			GetLogger().Info("请求完成", fields...)
		}
	}
}

// GinLogger 返回Gin的日志中间件
func GinLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 开始时间
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		// 处理请求
		c.Next()

		// 结束时间
		end := time.Now()
		latency := end.Sub(start)

		// 获取状态码和错误信息
		status := c.Writer.Status()
		clientIP := c.ClientIP()
		method := c.Request.Method

		// 构建日志字段
		fields := []zapcore.Field{
			zap.Int("status", status),
			zap.String("method", method),
			zap.String("path", path),
			zap.String("query", query),
			zap.String("ip", clientIP),
			zap.Duration("latency", latency),
			zap.String("user-agent", c.Request.UserAgent()),
		}

		// 根据状态码记录不同级别的日志
		switch {
		case status >= http.StatusInternalServerError:
			GetLogger().Error("服务器错误", fields...)
		case status >= http.StatusBadRequest:
			GetLogger().Warn("客户端错误", fields...)
		default:
			GetLogger().Info("请求完成", fields...)
		}
	}
}
