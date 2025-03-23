package config

import (
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
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
	Port        string
	Environment string
	Version     string
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
	TimeZone string
}

// JWTConfig JWT配置
type JWTConfig struct {
	Secret     string
	Expiration time.Duration
}

// EmailConfig 邮件配置
type EmailConfig struct {
	Host     string
	Port     string
	User     string
	Password string
}

// 全局配置实例
var AppConfig *Config

// LoadConfig 加载配置
func LoadConfig() *Config {
	// 加载环境变量
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using environment variables")
	}

	jwtExpiration, err := time.ParseDuration(getEnv("JWT_EXPIRATION", "24h"))
	if err != nil {
		log.Printf("Warning: Invalid JWT_EXPIRATION, using default 24h: %v", err)
		jwtExpiration = 24 * time.Hour
	}

	config := &Config{
		Server: ServerConfig{
			Port:        getEnv("PORT", "8080"),
			Environment: getEnv("ENV", "development"),
			Version:     "1.0.0",
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "postgres"),
			Name:     getEnv("DB_NAME", "codefolio"),
			SSLMode:  "disable",
			TimeZone: "Asia/Shanghai",
		},
		JWT: JWTConfig{
			Secret:     getEnv("JWT_SECRET", "your-secret-key"),
			Expiration: jwtExpiration,
		},
		Email: EmailConfig{
			Host:     getEnv("SMTP_HOST", "smtp.example.com"),
			Port:     getEnv("SMTP_PORT", "587"),
			User:     getEnv("SMTP_USER", ""),
			Password: getEnv("SMTP_PASSWORD", ""),
		},
	}

	AppConfig = config
	return config
}

// GetDSN 获取数据库连接字符串
func (c *Config) GetDSN() string {
	return "host=" + c.Database.Host +
		" user=" + c.Database.User +
		" password=" + c.Database.Password +
		" dbname=" + c.Database.Name +
		" port=" + c.Database.Port +
		" sslmode=" + c.Database.SSLMode +
		" TimeZone=" + c.Database.TimeZone
}

// getEnv 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
