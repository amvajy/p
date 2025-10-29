package api

import (
	"database/sql"
	"log"

	"pxe-manager/audit"
	"pxe-manager/auth"
	"pxe-manager/config"

	"github.com/gin-gonic/gin"
)

func SetupRouter(db *sql.DB, cfg *config.Config) *gin.Engine {
	r := gin.Default()

	// 反向代理信任列表（以确保正确获取 ClientIP）
	if err := r.SetTrustedProxies([]string{"127.0.0.1/32", "10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16"}); err != nil {
		log.Printf("配置 TrustedProxies 失败: %v", err)
	}

	// 开发环境 CORS，允许从 8000 静态预览访问 8080 API
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// 前端静态资源与模板（当使用 Go 服务时）
	r.Static("/static", "./web/static")
	r.LoadHTMLGlob("web/templates/*.html")
	r.GET("/", FrontendIndex())

	sec := &auth.SecurityConfig{
		AuthToken:    cfg.Auth.AuthToken,
		WhitelistIPs: cfg.Auth.WhitelistIPs,
		EnableAudit:  cfg.Auth.EnableAudit,
		AuditLogPath: cfg.Auth.AuditLogPath,
		RateLimit:    cfg.Auth.RateLimit,
	}

	// 限流中间件（白名单豁免）
	r.Use(auth.RateLimitMiddleware(sec))

	// 审计日志中间件（全局）
	var auditLogger *audit.AuditLogger
	if sec.EnableAudit {
		al, err := audit.NewAuditLogger(sec.AuditLogPath, true)
		if err != nil {
			log.Printf("初始化审计日志失败: %v", err)
		} else {
			auditLogger = al
			r.Use(AuditRequestMiddleware(auditLogger))
		}
	}

	apiGroup := r.Group("/api")
	apiGroup.Use(auth.Middleware(sec))

	// 健康检查
	apiGroup.GET("/health", HealthHandler())

	apiGroup.POST("/report", ReportHandler(db))
	apiGroup.GET("/servers", ListServersHandler(db))
	apiGroup.GET("/servers/:serial", GetServerHandler(db))
	apiGroup.POST("/servers/:serial/confirm", ConfirmServerHandler(db))
	apiGroup.POST("/servers/:serial/install", MarkInstalledHandler(db))

	apiGroup.GET("/configs", ListConfigsHandler(db))
	apiGroup.GET("/configs/:id", GetConfigHandler(db))
	apiGroup.POST("/configs", CreateConfigHandler(db))
	apiGroup.PUT("/configs/:id", UpdateConfigHandler(db))
	apiGroup.POST("/configs/:id/apply", ApplyConfigHandler(db, cfg))

	// 审计日志查看
	apiGroup.GET("/audit/logs", ListAuditLogsHandler(cfg))

	return r
}
