package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// AuthMiddleware JWT认证中间件
func AuthMiddleware(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			ResponseWithError(c, http.StatusUnauthorized, UNAUTHORIZED, "")
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			ResponseWithError(c, http.StatusUnauthorized, INVALID_TOKEN, "")
			c.Abort()
			return
		}

		tokenString := parts[1]
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return []byte(jwtSecret), nil
		})

		if err != nil || !token.Valid {
			ResponseWithError(c, http.StatusUnauthorized, INVALID_TOKEN, "")
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			ResponseWithError(c, http.StatusUnauthorized, INVALID_TOKEN, "")
			c.Abort()
			return
		}

		userID, ok := claims["user_id"].(float64)
		if !ok {
			ResponseWithError(c, http.StatusUnauthorized, INVALID_TOKEN, "")
			c.Abort()
			return
		}

		c.Set("user_id", uint(userID))
		c.Next()
	}
}

// CORSMiddleware 处理跨域请求的中间件
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
