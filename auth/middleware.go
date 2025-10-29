package auth

import (
	"net/http"
	"strings"

	"pxe-manager/utils"

	"github.com/gin-gonic/gin"
)

// Middleware 实现白名单优先 + Bearer 认证
func Middleware(sec *SecurityConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		// 白名单 IP 直接放行
		if utils.IsIPInWhitelist(clientIP, sec.WhitelistIPs) {
			c.Next()
			return
		}
		token := c.GetHeader("Authorization")
		if !strings.HasPrefix(token, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未提供认证令牌"})
			c.Abort()
			return
		}
		if strings.TrimPrefix(token, "Bearer ") != sec.AuthToken {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "无效的认证令牌"})
			c.Abort()
			return
		}
		c.Next()
	}
}
