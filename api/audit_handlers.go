package api

import (
	"bufio"
	"encoding/json"
	"os"
	"strconv"

	"pxe-manager/audit"
	"pxe-manager/config"

	"github.com/gin-gonic/gin"
)

// ListAuditLogsHandler 读取审计日志 JSON 行，支持 limit/offset/order
// GET /api/audit/logs?limit=100&offset=0&order=desc
func ListAuditLogsHandler(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		path := cfg.Auth.AuditLogPath
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
		if limit <= 0 {
			limit = 100
		}
		offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
		if offset < 0 {
			offset = 0
		}
		order := c.DefaultQuery("order", "desc") // desc: 最新在前

		f, err := os.Open(path)
		if err != nil {
			c.JSON(200, []audit.AuditLog{})
			return
		}
		defer f.Close()

		scanner := bufio.NewScanner(f)
		var lines []audit.AuditLog
		for scanner.Scan() {
			var item audit.AuditLog
			b := scanner.Bytes()
			if err := json.Unmarshal(b, &item); err == nil {
				lines = append(lines, item)
			}
		}
		// 根据 order/offset/limit 切片
		var out []audit.AuditLog
		if order == "desc" {
			for i := len(lines) - 1 - offset; i >= 0 && len(out) < limit; i-- {
				out = append(out, lines[i])
			}
		} else { // asc
			start := offset
			if start < 0 {
				start = 0
			}
			for i := start; i < len(lines) && len(out) < limit; i++ {
				out = append(out, lines[i])
			}
		}
		c.JSON(200, out)
	}
}
