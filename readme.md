📋 项目概述
基于Go语言开发的离线环境PXE管理系统，支持服务器信息收集、主机探测、装机配置管理等核心功能。

🏗️ 系统架构设计
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   客户端服务器    │    │   PXE管理服务器   │    │     数据库       │
│                 │    │                  │    │                 │
│ • 信息收集Agent  │───▶│ • API接收服务    │───▶│ • SQLite/MySQL  │
│ • 数据上报       │    │ • 数据确认管理    │    │ • 服务器信息     │
│                 │    │ • 配置管理       │    │ • 配置模板       │
└─────────────────┘    └──────────────────┘    └─────────────────┘
                                │
                                ▼
                       ┌──────────────────┐
                       │   TFTP/DHCP服务   │
                       │ • PXE启动文件    │
                       │ • 配置分发       │
                       └──────────────────┘
🗄️ 数据库设计
服务器信息表 (servers)
CREATE TABLE servers (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    serial VARCHAR(50) UNIQUE NOT NULL,
    hostname VARCHAR(100),
    ip_address VARCHAR(15),
    mac_address VARCHAR(17),
    gateway VARCHAR(15),
    install_time DATETIME,
    sda_size VARCHAR(20),
    part TEXT,
    system_version VARCHAR(100),
    kernel_version VARCHAR(100),
    cpu_model VARCHAR(100),
    cpu_processor INTEGER,
    mem_total INTEGER,
    memory_num INTEGER,
    lan_nic VARCHAR(50),
    lan_nic_speed VARCHAR(50),
    wan_nic VARCHAR(50),
    wan_nic_speed VARCHAR(50),
    bond_nic VARCHAR(50),
    bond_nic_speed VARCHAR(50),
    status VARCHAR(20) DEFAULT 'pending', -- pending, confirmed, installed
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

配置模板表 (config_templates)
CREATE TABLE config_templates (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    system_type VARCHAR(50), -- CentOS, Ubuntu, etc.
    system_version VARCHAR(50),
    config_content TEXT, -- Kickstart/Preseed配置
    kernel_params TEXT,
    packages TEXT, -- 预安装包列表
    status VARCHAR(20) DEFAULT 'active',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

🔧 Go模块结构
pxe-manager/
├── main.go                 # 程序入口
├── go.mod                 # 模块定义
├── config/
│   └── config.go          # 配置文件管理
├── api/
│   ├── handlers.go        # HTTP处理器
│   └── middleware.go      # 中间件
├── database/
│   ├── db.go             # 数据库连接
│   └── models.go         # 数据模型
├── services/
│   ├── server_service.go # 服务器管理服务
│   ├── config_service.go # 配置管理服务
│   └── pxe_service.go    # PXE服务管理
├── utils/
│   ├── network.go        # 网络工具
│   └── validation.go     # 数据验证
├── web/
│   ├── static/           # 静态文件
│   └── templates/        # HTML模板
└── scripts/
    └── agent/            # 客户端Agent脚本

🚀 核心功能实现
1. API服务端 (main.go)
package main

import (
    "log"
    "net/http"
    "pxe-manager/api"
    "pxe-manager/config"
    "pxe-manager/database"
)

func main() {
    // 加载配置
    cfg := config.LoadConfig()
    
    // 初始化数据库
    db, err := database.InitDB(cfg.DatabasePath)
    if err != nil {
        log.Fatal("数据库初始化失败:", err)
    }
    defer db.Close()
    
    // 设置路由
    router := api.SetupRouter(db, cfg)
    
    // 启动服务
    log.Printf("PXE管理系统启动在 %s", cfg.ServerAddress)
    log.Fatal(http.ListenAndServe(cfg.ServerAddress, router))
}

2. 配置管理 (config/config.go)
package config

type Config struct {
    ServerAddress string `json:"server_address"`
    DatabasePath  string `json:"database_path"`
    TFTPRoot      string `json:"tftp_root"`
    EnableDHCP    bool   `json:"enable_dhcp"`
    DHCPRange     string `json:"dhcp_range"`
    AuthToken     string `json:"auth_token"`
}

func LoadConfig() *Config {
    return &Config{
        ServerAddress: ":8080",
        DatabasePath:  "./data/pxe.db",
        TFTPRoot:      "/var/lib/tftpboot",
        EnableDHCP:    false,
        DHCPRange:     "192.168.88.100,192.168.88.200",
        AuthToken:     "your-secret-token",
    }
}

3. API处理器 (api/handlers.go)
package api

import (
    "encoding/json"
    "net/http"
    "pxe-manager/database"
    "pxe-manager/services"
)

type ServerReportRequest struct {
    Serial        string `json:"serial"`
    Hostname      string `json:"hostname"`
    IPAddress     string `json:"ip_address"`
    MACAddress    string `json:"mac_address"`
    Gateway       string `json:"gateway"`
    InstallTime   string `json:"install_time"`
    SdaSize       string `json:"sdaSize"`
    Part          string `json:"part"`
    SystemVersion string `json:"system_version"`
    KernelVersion string `json:"kernel_version"`
    CPUModel      string `json:"cpu_model"`
    CPUProcessor  int    `json:"cpu_processor"`
    MemTotal      int    `json:"MemTotal"`
    MemoryNum     int    `json:"Memory_num"`
    LanNic        string `json:"lanNic"`
    LanNicSpeed   string `json:"lanNic_Speed"`
    WanNic        string `json:"wanNic"`
    WanNicSpeed   string `json:"wanNic_Speed"`
    BondNic       string `json:"bondNic"`
    BondNicSpeed  string `json:"bondNic_Speed"`
}

func ReportHandler(db *database.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        var req ServerReportRequest
        if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
            http.Error(w, "无效的请求数据", http.StatusBadRequest)
            return
        }
        
        // 数据验证
        if req.Serial == "" || req.MACAddress == "" {
            http.Error(w, "序列号和MAC地址为必填项", http.StatusBadRequest)
            return
        }
        
        // 保存到数据库（待确认状态）
        server := database.Server{
            Serial:        req.Serial,
            Hostname:      req.Hostname,
            IPAddress:     req.IPAddress,
            MACAddress:    req.MACAddress,
            Gateway:       req.Gateway,
            InstallTime:   req.InstallTime,
            SdaSize:       req.SdaSize,
            Part:          req.Part,
            SystemVersion: req.SystemVersion,
            KernelVersion: req.KernelVersion,
            CPUModel:      req.CPUModel,
            CPUProcessor:  req.CPUProcessor,
            MemTotal:      req.MemTotal,
            MemoryNum:     req.MemoryNum,
            LanNic:        req.LanNic,
            LanNicSpeed:   req.LanNicSpeed,
            WanNic:        req.WanNic,
            WanNicSpeed:   req.WanNicSpeed,
            BondNic:       req.BondNic,
            BondNicSpeed:  req.BondNicSpeed,
            Status:        "pending",
        }
        
        if err := services.SaveServer(db, &server); err != nil {
            http.Error(w, "保存数据失败", http.StatusInternalServerError)
            return
        }
        
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(map[string]string{
            "status":  "success",
            "message": "服务器信息已接收，等待管理员确认",
        })
    }
}

// 确认服务器信息
func ConfirmServerHandler(db *database.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        serial := r.URL.Query().Get("serial")
        if serial == "" {
            http.Error(w, "缺少序列号参数", http.StatusBadRequest)
            return
        }
        
        if err := services.ConfirmServer(db, serial); err != nil {
            http.Error(w, "确认失败", http.StatusInternalServerError)
            return
        }
        
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(map[string]string{
            "status":  "success",
            "message": "服务器信息已确认",
        })
    }
}

4. Web配置编辑器 (web/templates/editor.html)
<!DOCTYPE html>
<html>
<head>
    <title>PXE配置编辑器</title>
    <style>
        .container { max-width: 1200px; margin: 0 auto; padding: 20px; }
        .server-list { border: 1px solid #ddd; margin-bottom: 20px; }
        .server-item { padding: 10px; border-bottom: 1px solid #eee; }
        .server-item.pending { background: #fff3cd; }
        .editor { display: flex; gap: 20px; }
        .config-list { width: 300px; }
        .config-editor { flex: 1; }
        textarea { width: 100%; height: 400px; font-family: monospace; }
    </style>
</head>
<body>
    <div class="container">
        <h1>PXE管理系统</h1>
        
        <div class="server-list" id="serverList">
            <h3>待确认服务器</h3>
        </div>
        
        <div class="editor">
            <div class="config-list">
                <h3>配置模板</h3>
                <select id="configSelect" onchange="loadConfig(this.value)">
                    <option value="">选择配置模板</option>
                </select>
                <button onclick="saveConfig()">保存配置</button>
            </div>
            
            <div class="config-editor">
                <h3>配置内容</h3>
                <textarea id="configContent"></textarea>
                <div>
                    <button onclick="applyConfig()">应用到服务器</button>
                    <select id="serverSelect"></select>
                </div>
            </div>
        </div>
    </div>

    <script>
        // 加载服务器列表
        async function loadServers() {
            const response = await fetch('/api/servers?status=pending');
            const servers = await response.json();
            
            const serverList = document.getElementById('serverList');
            const serverSelect = document.getElementById('serverSelect');
            
            servers.forEach(server => {
                const item = document.createElement('div');
                item.className = `server-item ${server.status}`;
                item.innerHTML = `
                    <strong>${server.serial}</strong> - ${server.hostname}
                    <button onclick="confirmServer('${server.serial}')">确认</button>
                `;
                serverList.appendChild(item);
                
                const option = document.createElement('option');
                option.value = server.serial;
                option.textContent = `${server.serial} - ${server.hostname}`;
                serverSelect.appendChild(option);
            });
        }
        
        // 确认服务器
        async function confirmServer(serial) {
            await fetch(`/api/servers/${serial}/confirm`, { method: 'POST' });
            loadServers();
        }
        
        // 加载配置模板
        async function loadConfigs() {
            const response = await fetch('/api/configs');
            const configs = await response.json();
            
            const select = document.getElementById('configSelect');
            configs.forEach(config => {
                const option = document.createElement('option');
                option.value = config.id;
                option.textContent = `${config.name} (${config.system_version})`;
                select.appendChild(option);
            });
        }
        
        // 加载具体配置内容
        async function loadConfig(configId) {
            if (!configId) return;
            
            const response = await fetch(`/api/configs/${configId}`);
            const config = await response.json();
            document.getElementById('configContent').value = config.config_content;
        }
        
        // 初始化
        loadServers();
        loadConfigs();
    </script>
</body>
</html>

5. 客户端Agent脚本 (scripts/agent/system_info.sh)
#!/bin/bash

# 系统信息收集脚本
SERVER_URL="http://pxe-manager:8080/api/report"
AUTH_TOKEN="your-auth-token"

collect_system_info() {
    # 基础信息
    SERIAL=$(dmidecode -s system-serial-number)
    HOSTNAME=$(hostname)
    IP_ADDRESS=$(ip route get 1 | awk '{print $7}')
    MAC_ADDRESS=$(ip link show eth0 | awk '/link\/ether/ {print $2}')
    GATEWAY=$(ip route | awk '/default/ {print $3}')
    
    # 磁盘信息
    SDA_SIZE=$(lsblk -b /dev/sda | awk 'NR==2 {print $4}')
    PART=$(lsblk -o NAME,SIZE,MOUNTPOINT /dev/sda | grep -v NAME | tr '\n' ',')
    
    # 系统信息
    SYSTEM_VERSION=$(cat /etc/redhat-release 2>/dev/null || cat /etc/os-release | grep PRETTY_NAME | cut -d'"' -f2)
    KERNEL_VERSION=$(uname -r)
    
    # CPU信息
    CPU_MODEL=$(grep "model name" /proc/cpuinfo | head -1 | cut -d':' -f2 | sed 's/^ *//')
    CPU_PROCESSOR=$(grep -c "^processor" /proc/cpuinfo)
    
    # 内存信息
    MEM_TOTAL=$(grep MemTotal /proc/meminfo | awk '{print $2}')
    MEMORY_NUM=$(dmidecode -t memory | grep "Size:" | grep -v "No Module" | wc -l)
    
    # 网络信息
    LAN_NIC="eth0"
    LAN_NIC_SPEED=$(ethtool eth0 2>/dev/null | grep Speed | awk '{print $2}' || echo "unknown")
    
    # 组装JSON数据
    JSON_DATA=$(cat <<EOF
{
    "serial": "$SERIAL",
    "hostname": "$HOSTNAME",
    "ip_address": "$IP_ADDRESS",
    "mac_address": "$MAC_ADDRESS",
    "gateway": "$GATEWAY",
    "install_time": "$(date '+%Y-%m-%d %H:%M:%S')",
    "sdaSize": "$SDA_SIZE",
    "part": "$PART",
    "system_version": "$SYSTEM_VERSION",
    "kernel_version": "$KERNEL_VERSION",
    "cpu_model": "$CPU_MODEL",
    "cpu_processor": $CPU_PROCESSOR,
    "MemTotal": $MEM_TOTAL,
    "Memory_num": $MEMORY_NUM,
    "lanNic": "$LAN_NIC",
    "lanNic_Speed": "$LAN_NIC_SPEED",
    "wanNic": "eth1",
    "wanNic_Speed": "unknown",
    "bondNic": "",
    "bondNic_Speed": ""
}
EOF
)

    # 发送数据到PXE管理系统
    curl -X POST \
         -H "Content-Type: application/json" \
         -H "Authorization: Bearer $AUTH_TOKEN" \
         -d "$JSON_DATA" \
         $SERVER_URL
}

# 执行信息收集
collect_system_info


📦 部署配置
编译配置 (go.mod)
module pxe-manager

go 1.19

require (
    github.com/gin-gonic/gin v1.9.1
    github.com/mattn/go-sqlite3 v1.14.17
    github.com/sirupsen/logrus v1.9.3
)

构建脚本 (build.sh)
#!/bin/bash

# 构建PXE管理系统
echo "构建PXE管理系统..."

# 编译Linux版本
GOOS=linux GOARCH=amd64 go build -o bin/pxe-manager-linux main.go

# 编译Windows版本  
GOOS=windows GOARCH=amd64 go build -o bin/pxe-manager-windows.exe main.go

echo "构建完成！"
echo "Linux版本: bin/pxe-manager-linux"
echo "Windows版本: bin/pxe-manager-windows.exe"



系统服务配置 (pxe-manager.service)
[Unit]
Description=PXE Management System
After=network.target

[Service]
Type=simple
User=pxe
WorkingDirectory=/opt/pxe-manager
ExecStart=/opt/pxe-manager/pxe-manager-linux
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target


🔧 核心特性
1. 主机探测功能
// 网络扫描发现新主机
func NetworkScan(subnet string) ([]Host, error) {
    // 实现ARP扫描或ICMP Ping扫描
    // 返回发现的MAC地址和IP地址
}

2. PXE配置生成
// 生成Kickstart配置
func GenerateKickstart(config *ConfigTemplate, server *Server) string {
    // 基于模板和服务器信息生成个性化安装配置
}

3. 配置版本管理
// 配置版本控制
type ConfigVersion struct {
    ID        int
    ConfigID  int
    Content   string
    Version   int
    CreatedAt time.Time
    CreatedBy string
}

🚀 部署步骤
编译二进制文件
./build.sh

部署到目标环境
# 创建数据目录
mkdir -p /opt/pxe-manager/data

# 复制二进制文件和配置文件
cp bin/pxe-manager-linux /opt/pxe-manager/
cp config.json /opt/pxe-manager/

# 设置服务
cp pxe-manager.service /etc/systemd/system/
systemctl daemon-reload
systemctl enable pxe-manager
systemctl start pxe-manager

配置TFTP服务
# 确保TFTP根目录存在
mkdir -p /var/lib/tftpboot/pxelinux.cfg

# 复制PXE启动文件
cp /usr/share/syslinux/pxelinux.0 /var/lib/tftpboot/

