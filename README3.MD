# PXE Manager 开发进展评估与下一步计划（README3）

## 一、当前进展概览
已完成端到端的“前端页面 ↔ 开发API桩 ↔ 预期Go后端”联调雏形，并在代码库中落地以下模块：
- 路由与中间件
  - 认证：白名单豁免 + Bearer Token 校验（`auth/middleware.go`）。
  - 限流：按 IP 每分钟限速，白名单不受限（`auth/ratelimit.go`）。
  - 审计：全局中间件记录 JSON 行日志（`api/audit_middleware.go` + `audit/logger.go`），浏览端点读取（`api/audit_handlers.go`）。
  - CORS：开发环境允许来自 8000 端口的前端访问 8080 API（`api/router.go`）。
- API 与静态前端
  - 健康检查：`GET /api/health`（`api/health.go`）。
  - 服务器：上报（桩端已实现，Go端入口在`api/handlers.go`中）、列表/详情、确认、标记安装。
  - 模板：列表、详情、创建、更新、对指定序列号应用并生成 PXE 配置（`api/handlers.go`、`pxe/*`）。
  - 审计日志：支持 `offset/limit/order` 的基本读取（`api/audit_handlers.go`）。
  - 静态/模板：`/static` 与 `web/templates/*.html`（`api/router.go`）。
- 数据库与迁移
  - 初始化：SQLite 使用 WAL、忙等待与外键启用；驱动已切换为 `modernc.org/sqlite`（无需 CGO），相关 PRAGMA 在启动时显式设置（`database/db.go`）。
  - 迁移：`servers/config_templates/processed_requests` 三表、索引、触发器完整（`migrations/001_initial.sql`）。
  - Repository：`SaveServer` UPSERT、状态变更、列表/详情、Config CRUD（`database/repository.go`）。
  - 模型映射：API camelCase ↔ DB snake_case（`database/models.go`）。
- PXE 生成
  - 生成器与模板：根据配置选择 Kickstart/Preseed，生成绑定文件（`pxe/generator.go`、`pxe/kickstart.go`、`pxe/preseed.go`）。
- 前端页面与交互
  - 首页：服务器列表（状态筛选、确认/安装/详情）、模板 CRUD、应用模板（`web/templates/index.html` + `web/static/js/app.js`）。
  - 审计：筛选与导出当前结果（方法/路径含/动作/状态）（`web/templates/audit.html` + `web/static/js/audit.js`）。
  - 详情页：确认、标记安装、应用模板（`web/templates/server.html` + `web/static/js/server.js`）。
- 运行环境
  - 前端预览端口：`8000`（Python http.server）。
  - 后端（Go）：`go run .` 或 `go run main.go`，默认监听 `:8080`；已替代开发API桩。

## 二、与 README2.MD 的对照与偏差
总体一致，关键点如下：
- 技术栈与结构：完全符合“Go+Gin、SQLite默认、白名单认证、限流、审计、PXE生成器”的架构设定。
- 配置结构：`config/config.go` 与 README2 示例一致（`server_address`、`database/sqlite_path`、`auth/token/ip_whitelist/rate_limit/enable_audit/audit_log_path`、`tftp/root/enable_uefi`）。其中 `AuthToken` 优先读取环境变量 `PXE_AUTH_TOKEN`，与说明一致。
- 数据模型与迁移：三张表、索引与触发器都已实现，WAL DSN 匹配 README2。
- 安全与认证：白名单免认证与免限流、Bearer Token校验、按 IP 限流，均符合。
- API设计：核心端点已实现；列表暂未分页与服务端筛选扩展（README2 标注“第二版加入”），目前前端与审计端支持客户端筛选。
- 字段规范：API camelCase 与 DB snake_case 映射已落地。
- PXE：绑定命名与引导文件路径策略已在生成器中实现默认；UEFI开关来自配置。
- 部署与运维：目前处于开发状态，尚未制作 Linux/Windows 二进制与 systemd 配置。
- 风险与边界：TrustedProxies 已配置；UEFI路径可扩展（默认在生成器中），后续可在配置结构中增加更细粒度字段。

可见偏差/待完善点：
- 开发运行：Go 后端已替换并运行；后续需补服务端过滤与分页。
- 审计后端过滤：`/api/audit/logs` 仅支持 `offset/limit/order`，未提供方法/路径/动作/状态的服务端过滤参数（前端已加客户端过滤）。
- 服务器列表：未实现分页与服务端多条件筛选（README2 第二版计划）。
- 安全增强：输入校验、结构化日志（logrus）尚未全面引入到处理器层。
- 部署：未提供构建产物、systemd 单元、TFTP目录准备脚本与检测。
- Agent：说明书要求的小驼峰字段与幂等 `requestId` 已在 Go 端处理器中支持，但实际 Agent 采集脚本尚未纳入或对齐（当前以桩为主）。

## 三、下一步详细计划
分阶段推进，优先补齐后端过滤与分页。

阶段A（已完成）
- 启动与联调
  - 在 Windows 开发环境安装 Go 工具链并运行 `go run .`（或 `go run main.go`），替换 Python 桩服务为实际 Go API。
  - 前端 `API_BASE` 指向 `http://localhost:8080/api`，验证 `/api/health`、`/api/servers`、`/api/configs`、`/api/audit/logs`。
- 配置与环境
  - 设置 `PXE_AUTH_TOKEN`；创建 `./data` 与 `./logs` 目录；确保审计日志路径可写。
- 数据库
  - 在 Windows 下使用 `modernc.org/sqlite`（无需 CGO），通过显式 PRAGMA 启用外键、忙等待与 WAL；CI 构建通过。

阶段B（功能完善，2–4天）
- 服务器列表服务端增强
  - `GET /api/servers` 支持分页（`limit/offset`）与条件（`status/hostnameContains/ipContains/macContains`）。
  - 数据库层增加相应 WHERE+LIMIT+OFFSET 构造；返回总数用于前端分页控件。
- 模板管理
  - 完善错误处理与字段校验，返回更明确的错误信息；引入 logrus 结构化日志。
- PXE生成
  - 在 `config.PXEConfig` 中增加可选字段（UEFI/Legacy 路径与文件名），默认保持当前值；在生成时校验目标目录存在。

阶段C（交互与部署，1周）
- 前端
  - 首页分页与搜索；详情页分栏展示与高亮关键字段；操作成功后统一“轻提示”。
  - 审计页面增加“服务端过滤开关”与“分页”，导出包含筛选条件。
- 部署
  - 提供 `Makefile` 或 `build.ps1`；编译 Linux/Windows 二进制。
  - `systemd` 单元示例与目录准备脚本；TFTP根目录引导文件校验与复制。
- 测试
  - 单元测试（Upsert、模板渲染、PXE生成）；集成测试（路由+认证+限流+审计）。

## 四、使用与运行（开发）
- 前端预览：`python -m http.server 8000 --directory web`。
- 后端（Go）：`go run .` 或 `go run main.go`，默认监听 `:8080`。
- 认证令牌：浏览器中输入与保存；仅非白名单IP需要。
- 审计日志：默认写入 `./logs/audit.log`，可通过审计页面查看与导出。

## 五、提交与版本控制
- 已初始化 Git 仓库并完成提交（包含前端、后端与迁移）。
- 已添加远程并推送至 GitHub，CI（`.github/workflows/go-ci.yml`）验证通过。

## 六、推送到 GitHub（Windows/PowerShell）
如需在新的环境中推送，可用以下两种方式之一：

1）使用 GitHub CLI（推荐）
- 安装后执行：
  - `gh auth login`（登录并授权）。
  - `gh repo create <your-org>/<repo-name> --private --source . --remote origin --push`。

2）使用 Git + 个人访问令牌（PAT）
- 在 GitHub 创建一个空仓库，例如 `https://github.com/<your-org>/<repo-name>.git`。
- 在 PowerShell 设置凭据：
  - `git remote add origin https://github.com/<your-org>/<repo-name>.git`
  - `git push -u origin master`
- 如需使用 PAT：在提示时输入用户名与 PAT（或将 PAT 嵌入 URL：`https://<username>:<PAT>@github.com/<org>/<repo>.git`）。

## 七、里程碑与验收对照
- 接收并保存上报数据，重复上报幂等（已实现 `processed_requests` 主键与清理触发器；桩端已模拟）。
- 管理端查看/确认/模板维护，生成并落地 PXE 配置（已实现生成器与端点；需联通真实 TFTP 根目录）。
- SQLite WAL/触发器/索引（已实现）；限流与白名单（已实现）；审计落盘（已实现）。
- CI：GitHub Actions 已通过构建（Windows 无 CGO）。

## 八、风险与缓解
- 离线网络中的真实IP获取：已设置 TrustedProxies；仍需在反向代理处配置以确保 `ClientIP` 正确。
- Windows构建：`modernc.org/sqlite` 无需 CGO，减少工具链依赖；外键/忙等待/WAL 通过 PRAGMA 显式设置。
- IDE/终端网络：在部分环境中终端到本地环回端口的访问可能受限，建议用浏览器直接访问 `http://localhost:8080/api/health` 验证。
- UEFI路径大小写：生成器内置默认，后续将通过配置字段确保可变更。

——
如需我继续完善服务端过滤、分页与构建脚本，请告知优先级与目标平台。