package config

import (
	"encoding/json"
	"log"
	"os"
)

// 数据库配置
type DBConfig struct {
	Driver     string `json:"driver"`      // sqlite/mysql
	SQLitePath string `json:"sqlite_path"`
	MySQLDSN   string `json:"mysql_dsn"`
}

// 安全配置
type SecurityConfig struct {
	AuthToken    string   `json:"-"`             // 从环境变量读取优先
	WhitelistIPs []string `json:"ip_whitelist"`  // CIDR 或 IP
	RateLimit    int      `json:"rate_limit"`    // 每IP每分钟请求数
	EnableAudit  bool     `json:"enable_audit"`
	AuditLogPath string   `json:"audit_log_path"`
}

// PXE/TFTP 配置
type PXEConfig struct {
	Root       string `json:"root"`
	EnableUEFI bool   `json:"enable_uefi"`
}

// 顶层配置
type Config struct {
	ServerAddress string         `json:"server_address"`
	Database      DBConfig       `json:"database"`
	Auth          SecurityConfig `json:"auth"`
	TFTP          PXEConfig      `json:"tftp"`
}

func LoadConfig() *Config {
	// 尝试读取配置文件（可选）。未找到则使用默认值。
	cfg := &Config{
		ServerAddress: ":8080",
		Database: DBConfig{
			Driver:     "sqlite",
			SQLitePath: "./data/pxe.db",
		},
		Auth: SecurityConfig{
			WhitelistIPs: []string{"192.168.88.0/24"},
			RateLimit:    100,
			EnableAudit:  true,
			AuditLogPath: "./logs/audit.log",
		},
		TFTP: PXEConfig{
			Root:       "/var/lib/tftpboot",
			EnableUEFI: true,
		},
	}

	// 如果存在 config.json 则覆盖默认值
	if _, err := os.Stat("config.json"); err == nil {
		f, err := os.Open("config.json")
		if err == nil {
			defer f.Close()
			dec := json.NewDecoder(f)
			if err := dec.Decode(cfg); err != nil {
				log.Printf("读取配置文件失败，使用默认配置: %v", err)
			}
		}
	}

	// 环境变量覆盖认证令牌
	if token := os.Getenv("PXE_AUTH_TOKEN"); token != "" {
		cfg.Auth.AuthToken = token
	}

	return cfg
}
