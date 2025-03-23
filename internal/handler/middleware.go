package handler

import (
	"codefolio/internal/common"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// CORSMiddleware 处理跨域请求
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
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			common.ResponseWithError(c, common.CodeUnauthorized, http.StatusUnauthorized)
			c.Abort()
			return
		}

		// 检查认证头格式
		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			common.ResponseWithError(c, common.CodeInvalidToken, http.StatusUnauthorized)
			c.Abort()
			return
		}

		// 解析JWT令牌
		token, err := jwt.Parse(parts[1], func(token *jwt.Token) (interface{}, error) {
			// 验证签名算法
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(jwtSecret), nil
		})

		if err != nil {
			if err == jwt.ErrTokenExpired {
				common.ResponseWithError(c, common.CodeTokenExpired, http.StatusUnauthorized)
			} else {
				common.ResponseWithError(c, common.CodeInvalidToken, http.StatusUnauthorized)
			}
			c.Abort()
			return
		}

		// 验证令牌有效性
		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			// 检查令牌是否过期
			if exp, ok := claims["exp"].(float64); ok {
				if time.Now().Unix() > int64(exp) {
					common.ResponseWithError(c, common.CodeTokenExpired, http.StatusUnauthorized)
					c.Abort()
					return
				}
			}

			// 将用户ID添加到上下文
			if userID, ok := claims["sub"].(string); ok {
				c.Set("userID", userID)
			} else {
				common.ResponseWithError(c, common.CodeInvalidToken, http.StatusUnauthorized)
				c.Abort()
				return
			}
		} else {
			common.ResponseWithError(c, common.CodeInvalidToken, http.StatusUnauthorized)
			c.Abort()
			return
		}

		c.Next()
	}
}
