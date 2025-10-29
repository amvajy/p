package auth

// 安全配置结构（从 config.Config 中传入）
type SecurityConfig struct {
	AuthToken    string
	WhitelistIPs []string
	EnableAudit  bool
	AuditLogPath string
	RateLimit    int // 每IP每分钟请求数
}
