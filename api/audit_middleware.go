package api

import (
	"pxe-manager/audit"
	"github.com/gin-gonic/gin"
)

// AuditRequestMiddleware 记录每个请求的审计日志（通用请求层）
func AuditRequestMiddleware(logger *audit.AuditLogger) gin.HandlerFunc {
	return func(c *gin.Context) {
		if logger == nil {
			c.Next()
			return
		}
		// 将审计记录器存放在上下文，供处理器使用
		c.Set("auditLogger", logger)
		c.Next()
		status := "success"
		if c.Writer.Status() >= 400 {
			status = "failure"
		}
		_ = logger.LogEvent(audit.AuditLog{
			ClientIP:  c.ClientIP(),
			UserAgent: c.GetHeader("User-Agent"),
			Method:    c.Request.Method,
			Path:      c.Request.URL.Path,
			Action:    "http_request",
			Target:    c.Request.URL.Path,
			Status:    status,
		})
	}
}
