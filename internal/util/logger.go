package util

import (
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Logger *zap.Logger

// InitLogger 初始化日志
func InitLogger(env string) *zap.Logger {
	var config zap.Config

	if env == "production" {
		// 生产环境: 使用JSON格式，写入到文件
		config = zap.NewProductionConfig()
		config.OutputPaths = []string{"logs/app.log", "stdout"}
		config.ErrorOutputPaths = []string{"logs/error.log", "stderr"}
	} else {
		// 开发环境: 彩色控制台输出
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	// 公共配置
	config.EncoderConfig.TimeKey = "time"
	config.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(time.RFC3339)

	// 创建日志目录
	if env == "production" {
		os.MkdirAll("logs", 0755)
	}

	// 构建日志
	logger, err := config.Build()
	if err != nil {
		panic("Failed to initialize logger: " + err.Error())
	}

	// 替换全局日志
	zap.ReplaceGlobals(logger)

	// 存储日志实例
	Logger = logger

	return logger
}

// LoggerMiddleware Gin日志中间件
func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method
		ip := c.ClientIP()

		// 处理请求
		c.Next()

		// 请求处理完成，记录信息
		latency := time.Since(start)
		status := c.Writer.Status()

		// 确定日志级别
		var logFunc func(string, ...zap.Field)
		if status >= 500 {
			logFunc = Logger.Error
		} else if status >= 400 {
			logFunc = Logger.Warn
		} else {
			logFunc = Logger.Info
		}

		// 记录请求日志
		logFunc("HTTP Request",
			zap.String("method", method),
			zap.String("path", path),
			zap.Int("status", status),
			zap.Duration("latency", latency),
			zap.String("ip", ip),
			zap.Int("size", c.Writer.Size()),
			zap.String("user-agent", c.Request.UserAgent()),
		)
	}
}
