-- servers 表
CREATE TABLE IF NOT EXISTS servers (
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
    status VARCHAR(20) DEFAULT 'pending',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_servers_serial ON servers(serial);
CREATE INDEX IF NOT EXISTS idx_servers_mac ON servers(mac_address);
CREATE INDEX IF NOT EXISTS idx_servers_status ON servers(status);

-- updated_at 触发器（列集限定 AFTER UPDATE，避免递归）
CREATE TRIGGER IF NOT EXISTS update_servers_timestamp
AFTER UPDATE OF serial, hostname, ip_address, mac_address, gateway, install_time,
                 sda_size, part, system_version, kernel_version, cpu_model,
                 cpu_processor, mem_total, memory_num, lan_nic, lan_nic_speed,
                 wan_nic, wan_nic_speed, bond_nic, bond_nic_speed, status
ON servers
FOR EACH ROW
BEGIN
    UPDATE servers SET updated_at = CURRENT_TIMESTAMP WHERE id = OLD.id;
END;

-- config_templates 表
CREATE TABLE IF NOT EXISTS config_templates (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    system_type VARCHAR(50),
    system_version VARCHAR(50),
    config_content TEXT,
    kernel_params TEXT,
    packages TEXT,
    status VARCHAR(20) DEFAULT 'active',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- 幂等请求表
CREATE TABLE IF NOT EXISTS processed_requests (
    serial VARCHAR(50) NOT NULL,
    request_id VARCHAR(64) NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (serial, request_id)
);

CREATE INDEX IF NOT EXISTS idx_processed_requests_created_at ON processed_requests(created_at);

-- 清理 72 小时前的幂等记录
CREATE TRIGGER IF NOT EXISTS cleanup_processed_requests
AFTER INSERT ON processed_requests
BEGIN
    DELETE FROM processed_requests WHERE created_at < datetime('now', '-72 hours');
END;
