package handler

import (
	"codefolio/internal/common"
	"codefolio/internal/util"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"go.uber.org/zap"
)

// AuthClaims JWT声明结构
type AuthClaims struct {
	UserID string `json:"user_id"`
	jwt.StandardClaims
}

// Recovery 恢复中间件，将捕获的panic转换为500错误响应
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				util.GetLogger().Error("服务器发生panic",
					zap.Any("error", err),
					zap.String("request", c.Request.URL.Path))

				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"code":    500,
					"message": "服务器内部错误",
				})
			}
		}()
		c.Next()
	}
}

// LoggerMiddleware 日志中间件，记录HTTP请求
func LoggerMiddleware() gin.HandlerFunc {
	return util.GinLogger()
}

// CORSMiddleware 处理跨域请求中间件
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// AuthMiddleware JWT认证中间件
func AuthMiddleware(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取Authorization头
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			common.ResponseWithError(c, common.CodeUnauthorized, http.StatusUnauthorized)
			c.Abort()
			return
		}

		// 提取令牌
		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			common.ResponseWithError(c, common.CodeUnauthorized, http.StatusUnauthorized)
			c.Abort()
			return
		}

		// 解析JWT
		tokenString := parts[1]
		token, err := jwt.ParseWithClaims(tokenString, &AuthClaims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(jwtSecret), nil
		})

		// 检查解析错误
		if err != nil {
			util.GetLogger().Error("JWT解析失败", zap.Error(err))
			common.ResponseWithError(c, common.CodeUnauthorized, http.StatusUnauthorized)
			c.Abort()
			return
		}

		// 检查令牌有效性
		if !token.Valid {
			util.GetLogger().Error("无效的JWT")
			common.ResponseWithError(c, common.CodeUnauthorized, http.StatusUnauthorized)
			c.Abort()
			return
		}

		// 验证令牌是否过期
		claims, ok := token.Claims.(*AuthClaims)
		if !ok {
			util.GetLogger().Error("JWT声明类型无效")
			common.ResponseWithError(c, common.CodeUnauthorized, http.StatusUnauthorized)
			c.Abort()
			return
		}

		// 检查令牌是否过期
		if claims.ExpiresAt < time.Now().Unix() {
			util.GetLogger().Error("JWT已过期")
			common.ResponseWithError(c, common.CodeUnauthorized, http.StatusUnauthorized)
			c.Abort()
			return
		}

		// 将用户ID设置到上下文
		c.Set("userID", claims.UserID)
		c.Next()
	}
}
