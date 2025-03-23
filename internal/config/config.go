package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"go.uber.org/zap"
)

// Config 应用程序配置
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
	Email    EmailConfig
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Host        string
	Port        string
	Environment string
	Version     string
	LogDir      string
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Host     string
	Port     string
	Name     string
	User     string
	Password string
	SSLMode  string
	TimeZone string
}

// JWTConfig JWT配置
type JWTConfig struct {
	Secret        string
	ExpiryHours   int
	RefreshSecret string
}

// EmailConfig 邮件配置
type EmailConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
}

// LoadConfig 加载应用程序配置
func LoadConfig() *Config {
	// 尝试加载.env文件
	dotenvPath := ".env"
	if _, err := os.Stat(dotenvPath); os.IsNotExist(err) {
		// 如果.env文件不存在，记录警告但继续执行
		fmt.Println("警告: .env文件不存在，使用环境变量或默认值")
	}

	config := &Config{
		Server: ServerConfig{
			Host:        getEnv("SERVER_HOST", "localhost"),
			Port:        getEnv("SERVER_PORT", "8080"),
			Environment: getEnv("ENV", "development"),
			Version:     getEnv("VERSION", "1.0.0"),
			LogDir:      getEnv("LOG_DIR", "logs"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			Name:     getEnv("DB_NAME", "codefolio"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "postgres"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
			TimeZone: getEnv("DB_TIMEZONE", "Asia/Shanghai"),
		},
		JWT: JWTConfig{
			Secret:        getEnv("JWT_SECRET", "your-secret-key"),
			ExpiryHours:   getEnvAsInt("JWT_EXPIRY_HOURS", 24),
			RefreshSecret: getEnv("JWT_REFRESH_SECRET", "your-refresh-secret-key"),
		},
		Email: EmailConfig{
			Host:     getEnv("EMAIL_HOST", "smtp.example.com"),
			Port:     getEnvAsInt("EMAIL_PORT", 587),
			Username: getEnv("EMAIL_USERNAME", ""),
			Password: getEnv("EMAIL_PASSWORD", ""),
			From:     getEnv("EMAIL_FROM", "noreply@example.com"),
		},
	}

	return config
}

// GetDSN 获取数据库连接字符串
func (c *Config) GetDSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=%s",
		c.Database.Host,
		c.Database.Port,
		c.Database.User,
		c.Database.Password,
		c.Database.Name,
		c.Database.SSLMode,
		c.Database.TimeZone,
	)
}

// GetJWTExpiryDuration 获取JWT过期时间
func (c *Config) GetJWTExpiryDuration() time.Duration {
	return time.Duration(c.JWT.ExpiryHours) * time.Hour
}

// GetLogger 获取日志记录器
func (c *Config) GetLogger() (*zap.Logger, error) {
	var logger *zap.Logger
	var err error

	if c.Server.Environment == "production" {
		config := zap.NewProductionConfig()
		config.OutputPaths = []string{
			"stdout",
			fmt.Sprintf("%s/app.log", c.Server.LogDir),
		}
		logger, err = config.Build()
	} else {
		config := zap.NewDevelopmentConfig()
		logger, err = config.Build()
	}

	if err != nil {
		return nil, err
	}

	return logger, nil
}

// 辅助函数: 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// 辅助函数: 获取环境变量并转换为整数，如果不存在或转换失败则返回默认值
func getEnvAsInt(key string, defaultValue int) int {
	valueStr := getEnv(key, "")
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}

	return value
}
