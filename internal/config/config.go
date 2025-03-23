package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

// Config 应用配置结构
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
	Email    EmailConfig
	Upload   UploadConfig
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port    int
	Mode    string
	Version string
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	Name     string
	TimeZone string
}

// JWTConfig JWT配置
type JWTConfig struct {
	Secret      string
	ExpireHours int
}

// EmailConfig 邮件配置
type EmailConfig struct {
	SMTPHost string
	SMTPPort int
	Username string
	Password string
	From     string
}

// UploadConfig 文件上传配置
type UploadConfig struct {
	MaxFileSize   int64  // 最大文件大小（字节）
	AllowedTypes  string // 允许的文件类型
	StoragePath   string // 存储路径
	AnonymousView int    // 匿名用户查看限制
	UserView      int    // 注册用户查看限制
}

// LoadConfig 加载配置
func LoadConfig() *Config {
	// 尝试从.env文件加载环境变量
	if err := godotenv.Load(); err != nil {
		zap.L().Warn("未找到.env文件，将使用环境变量")
	}

	return &Config{
		Server: ServerConfig{
			Port:    getEnvAsInt("SERVER_PORT", 8080),
			Mode:    getEnv("SERVER_MODE", "development"),
			Version: getEnv("SERVER_VERSION", "0.1.0"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnvAsInt("DB_PORT", 3306),
			Username: getEnv("DB_USER", "root"),
			Password: getEnv("DB_PASSWORD", ""),
			Name:     getEnv("DB_NAME", "codefolio"),
			TimeZone: getEnv("DB_TIMEZONE", "Asia/Shanghai"),
		},
		JWT: JWTConfig{
			Secret:      getEnv("JWT_SECRET", "your-secret-key"),
			ExpireHours: getEnvAsInt("JWT_EXPIRE_HOURS", 24*7), // 默认7天
		},
		Email: EmailConfig{
			SMTPHost: getEnv("EMAIL_SMTP_HOST", ""),
			SMTPPort: getEnvAsInt("EMAIL_SMTP_PORT", 587),
			Username: getEnv("EMAIL_USERNAME", ""),
			Password: getEnv("EMAIL_PASSWORD", ""),
			From:     getEnv("EMAIL_FROM", ""),
		},
		Upload: UploadConfig{
			MaxFileSize:   getEnvAsInt64("UPLOAD_MAX_SIZE", 10*1024*1024), // 默认10MB
			AllowedTypes:  getEnv("UPLOAD_ALLOWED_TYPES", ".pdf"),
			StoragePath:   getEnv("UPLOAD_STORAGE_PATH", "./uploads"),
			AnonymousView: getEnvAsInt("UPLOAD_ANONYMOUS_VIEW_LIMIT", 5), // 匿名用户每天可查看5份简历
			UserView:      getEnvAsInt("UPLOAD_USER_VIEW_LIMIT", 20),     // 注册用户每天可查看20份简历
		},
	}
}

// GetDSN 获取数据库连接字符串
func (c *Config) GetDSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=%s",
		c.Database.Username,
		c.Database.Password,
		c.Database.Host,
		c.Database.Port,
		c.Database.Name,
		c.Database.TimeZone,
	)
}

// getEnv 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// getEnvAsInt 获取环境变量并转换为整数
func getEnvAsInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getEnvAsInt64 获取环境变量并转换为int64
func getEnvAsInt64(key string, defaultValue int64) int64 {
	if value, exists := os.LookupEnv(key); exists {
		if intValue, err := strconv.ParseInt(value, 10, 64); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getEnvAsBool 获取环境变量并转换为布尔值
func getEnvAsBool(key string, defaultValue bool) bool {
	if value, exists := os.LookupEnv(key); exists {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

// getEnvAsDuration 获取环境变量并转换为时间间隔
func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	if value, exists := os.LookupEnv(key); exists {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
