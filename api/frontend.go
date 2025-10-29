package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// FrontendIndex 渲染前端首页（当使用 Go 服务预览时）
func FrontendIndex() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Cache-Control", "no-cache")
		c.HTML(http.StatusOK, "index.html", gin.H{})
	}
}
