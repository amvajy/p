项目需求说明书（PXE 管理系统）

一、项目目标与范围
- 目标：在离线/隔离网络中集中管理 PXE 装机，完成服务器信息采集、确认、模板管理与 PXE 引导配置分发。
- 范围：后端 API、Web 管理前端、PXE 配置生成（Legacy/UEFI）、Agent 采集脚本、部署与审计。
- 不做：第一版不托管 DHCP，仅生成供外部集成的配置片段。

二、系统架构
- 组件：客户端 Agent → API 服务 → 数据库（默认 SQLite） → PXE/TFTP（文件分发） → Web 前端。
- 状态流转：pending（上报）→ confirmed（已确认）→ installed（装机完成，第一版手动标记）。

三、技术栈与运行环境
- 后端：Go 1.19+、Gin、modernc.org/sqlite（无CGO）、logrus（后续引入）。
- 数据库：默认 SQLite，保留 MySQL 扩展。
- 前端：原生 HTML/JS（先完善基础交互）。
- 运行：Linux 优先，Windows 构建主要用于开发/编译；TFTP 在 Windows 下可选启用并预警。

四、配置结构（示例）
{
  "server_address": ":8080",
  "database": {
    "driver": "sqlite",
    "sqlite_path": "./data/pxe.db"
  },
  "auth": {
    "token": "your-secure-token",
    "ip_whitelist": ["192.168.88.0/24"],
    "rate_limit": 100,
    "enable_audit": true,
    "audit_log_path": "./logs/audit.log"
  },
  "tftp": {
    "root": "/var/lib/tftpboot",
    "enable_uefi": true
  }
}

五、数据模型与迁移
- 表 servers：字段含 serial（唯一）、hostname、ip_address、mac_address、硬件与状态、created_at、updated_at。
- 表 config_templates：模板内容、系统类型版本、kernel_params、packages、status。
- 表 processed_requests：serial、request_id、created_at（主键：serial+request_id，幂等去重，72 小时清理）。
- 索引：serial（唯一）、mac_address、status；processed_requests.created_at。
- 触发器：在 servers 上使用“列集限定的 AFTER UPDATE”触发器，将 updated_at 更新为 CURRENT_TIMESTAMP。
- SQLite 连接：驱动 `sqlite`，初始化时通过显式 PRAGMA 设置 foreign_keys=ON、busy_timeout=5000、journal_mode=WAL。

六、安全与认证
- 认证：Bearer Token（环境变量 `PXE_AUTH_TOKEN` 优先，开发环境可回退配置）。
- 白名单：支持单 IP 与 CIDR，白名单 IP 免认证与免限流。
- 审计：EnableAudit=true 时写入 JSON 行日志到 `audit_log_path`，记录时间、IP、路由、动作、目标与结果。

七、速率限制
- 口径：每 IP 每分钟请求数，默认 100 次/分钟。
- 维度：按 IP 维护独立限流器（本地缓存，带 TTL 清理）；非白名单生效。

八、API 设计（核心）
- POST `/api/report`：Agent 上报；含 `requestId` 幂等；返回接收状态。
- GET `/api/servers?status=pending`：列表；支持分页与筛选（第二版加入）。
- GET `/api/servers/:serial`：详情。
- POST `/api/servers/:serial/confirm`：确认服务器为 confirmed。
- POST `/api/servers/:serial/install`：标记为 installed（第一版手动）。
- GET `/api/configs`：模板列表；GET `/api/configs/:id`：模板详情。
- POST `/api/configs`：创建模板；PUT `/api/configs/:id`：更新模板。
- POST `/api/configs/:id/apply?serial=...`：应用模板并为目标服务器生成 PXE 绑定配置。
- 认证：除白名单外，所有 `/api/*` 需 `Authorization: Bearer <token>`。

九、字段规范（API/DB）
- API 使用 camelCase，DB 使用 snake_case，通过 `json/db` tag 映射。
- Server（核心字段）：serial、hostname、ipAddress、macAddress、gateway、installTime、sdaSize、part、systemVersion、kernelVersion、cpuModel、cpuProcessor、memTotal、memoryNum、lanNic、lanNicSpeed、wanNic、wanNicSpeed、bondNic、bondNicSpeed、status。
- ConfigTemplate：name、description、systemType、systemVersion、configContent、kernelParams、packages、status。

十、PXE 引导与配置
- 绑定命名：`pxelinux.cfg/01-aa-bb-cc-dd-ee-ff`（小写连字符，MAC 去 `:`）。
- Legacy（Syslinux）：引导文件 `pxelinux.0`，配置路径 `pxelinux.cfg/`。
- UEFI（GRUB）：引导文件 `grubx64.efi`，配置文件 `grub.cfg`，可自定义路径（如 `efi/boot/`）。
- 生成器：根据 `systemType` 选择 Kickstart（CentOS/RHEL）或 Preseed（Ubuntu/Debian），渲染模板并输出至 TFTP 目录（确保目录存在）。

十一、Agent 设计
- 采集：序列号（root 用 dmidecode，降级用 machine-id/主机名）、主网卡自适应（默认路由/遍历 `/sys/class/net`）、CPU/内存/磁盘信息、系统与内核版本。
- 字段：统一小驼峰；生成 `requestId`（UUID 或替代）；支持超时与最多 3 次重试；离线容错与日志提示。
- 环境：`SERVER_URL`、`AUTH_TOKEN`，建议通过配置或环境变量注入。

十二、Web 前端（第一版）
- 页面：服务器确认与模板编辑器（基础 CRUD）。
- 交互：加载列表前清空容器；完善 `saveConfig()` 与 `applyConfig()`；错误处理与加载态；成功提示。
- 第二版：分页、搜索与实时状态；第三版：权限分级与 WebSocket。

十三、部署与运维
- 构建：提供 Linux/Windows 二进制（Windows 主要用于开发/编译）。
- Linux 部署：systemd 管理；创建数据目录与复制可执行文件及配置；TFTP 根目录准备与引导文件复制。
- 安全：AuthToken 从环境变量读取；启用审计日志；限制对外暴露端口。

十四、质量保障与测试
- 单元测试：服务层（Upsert、模板渲染、PXE 生成）。
- 集成测试：API 路由与认证、限流、白名单、审计日志。
- 运行验证：SQLite WAL 生效、触发器更新 `updated_at`、TFTP 目录生成文件。

十五、实施计划
- 第一阶段（1–2 周）：数据库与模型、认证与路由、PXE 生成与 TFTP 分发、基础限流与审计。
- 第二阶段（1 周）：Agent 改进、前端交互完善、稳定性增强。
- 第三阶段（1 周）：部署优化、结构化日志、安全加固（IP 白名单、输入校验、CORS）。

十六、边界与风险
- 离线环境权限（dmidecode 需 root）与网卡命名差异；通过降级兼容。
- 反向代理获取真实 IP（需配置 TrustedProxies），影响限流与白名单。
- UEFI 路径大小写可能随环境变化，支持自定义并给出默认值。

十七、验收标准
- 能接收并保存上报数据，重复上报幂等。
- 管理端可查看/确认服务器与维护模板，能生成并落地 MAC 绑定的 PXE 配置。
- SQLite WAL、触发器与索引生效；限流与白名单按策略工作；审计日志落盘。

十八、变更控制
- 所有功能实现与结构调整在开发前需你明确同意；重大改动（数据结构、认证策略、生成路径）单独评审与确认。