package handler

import (
	"log"
	"net/http"
	"strings"

	"cyber-inspector/internal/auth"
	"github.com/gin-gonic/gin"
)

// AuthMiddleware 认证中间件（带调试日志）
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		log.Printf("[JWTDebug] 原始头: %s", authHeader)
		if authHeader == "" {
			log.Printf("[JWTDebug] 空Authorization头")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未提供访问令牌"})
			c.Abort()
			return
		}

		// 去掉 Bearer 前缀
		if len(authHeader) > 7 && strings.HasPrefix(authHeader, "Bearer ") {
			authHeader = authHeader[7:]
		} else {
			log.Printf("[JWTDebug] 非Bearer格式")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "无效的访问令牌"})
			c.Abort()
			return
		}

		log.Printf("[JWTDebug] 待验token: %s", authHeader)

		// 解析 Token
		claims, err := auth.ParseToken(authHeader)
		if err != nil {
			log.Printf("[JWTDebug] 解析失败: %v", err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "无效的访问令牌"})
			c.Abort()
			return
		}

		// 将用户信息存入上下文
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)
		c.Next()
	}
}

// AdminMiddleware 管理员权限中间件
func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists || role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "需要管理员权限"})
			c.Abort()
			return
		}
		c.Next()
	}
}
