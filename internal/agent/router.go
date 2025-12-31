package agent

import (
	"fmt"
	"github.com/gin-gonic/gin"
)

// RequestLogger 返回一个 Gin 日志中间件
func RequestLogger() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("[%s] %s %s %d %v\n",
			param.TimeStamp.Format("2006-01-02 15:04:05"),
			param.Method,
			param.Path,
			param.StatusCode,
			param.Latency,
		)
	})
}

// RegisterRoutes 注册 Agent 相关路由
func RegisterRoutes(r *gin.Engine) {
	// 心跳接口，后面再补其它
	r.GET("/ping", func(c *gin.Context) {
		c.String(200, "pong")
	})
}
