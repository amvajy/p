package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// HealthHandler 简单健康检查端点，用于前端连通性探测
func HealthHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	}
}
