package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"pxe-manager/config"
	"pxe-manager/database"
	"pxe-manager/pxe"
	"pxe-manager/audit"

	"github.com/gin-gonic/gin"
)

type ServerReportRequest struct {
	RequestID     string `json:"requestId"`
	Serial        string `json:"serial"`
	Hostname      string `json:"hostname"`
	IPAddress     string `json:"ipAddress"`
	MACAddress    string `json:"macAddress"`
	Gateway       string `json:"gateway"`
	InstallTime   string `json:"installTime"`
	SdaSize       string `json:"sdaSize"`
	Part          string `json:"part"`
	SystemVersion string `json:"systemVersion"`
	KernelVersion string `json:"kernelVersion"`
	CPUModel      string `json:"cpuModel"`
	CPUProcessor  int    `json:"cpuProcessor"`
	MemTotal      int    `json:"memTotal"`
	MemoryNum     int    `json:"memoryNum"`
	LanNic        string `json:"lanNic"`
	LanNicSpeed   string `json:"lanNicSpeed"`
	WanNic        string `json:"wanNic"`
	WanNicSpeed   string `json:"wanNicSpeed"`
	BondNic       string `json:"bondNic"`
	BondNicSpeed  string `json:"bondNicSpeed"`
}

// auditEvent 将动作级别的审计事件写入日志（若可用）
func auditEvent(c *gin.Context, action, target, status string) {
	v, ok := c.Get("auditLogger")
	if !ok || v == nil {
		return
	}
	logger, ok := v.(*audit.AuditLogger)
	if !ok || logger == nil {
		return
	}
	_ = logger.LogEvent(audit.AuditLog{
		ClientIP:  c.ClientIP(),
		UserAgent: c.GetHeader("User-Agent"),
		Method:    c.Request.Method,
		Path:      c.Request.URL.Path,
		Action:    action,
		Target:    target,
		Status:    status,
	})
}

func ReportHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req ServerReportRequest
		if err := c.BindJSON(&req); err != nil {
			auditEvent(c, "report", "", "failure")
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求数据"})
			return
		}
		if req.Serial == "" || req.MACAddress == "" || req.RequestID == "" {
			auditEvent(c, "report", req.Serial, "failure")
			c.JSON(http.StatusBadRequest, gin.H{"error": "缺少必填字段(serial/macAddress/requestId)"})
			return
		}
		_, err := db.Exec(`INSERT INTO processed_requests(serial, request_id) VALUES(?,?)`, req.Serial, req.RequestID)
		if err != nil {
			// 如果主键冲突，视为已处理
			auditEvent(c, "report", req.Serial, "duplicate")
			c.JSON(http.StatusOK, gin.H{"status": "success", "message": "重复请求，已忽略"})
			return
		}

		s := &database.Server{
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
		if err := database.SaveServer(db, s); err != nil {
			auditEvent(c, "report", req.Serial, "failure")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "保存数据失败"})
			return
		}
		auditEvent(c, "report", req.Serial, "success")
		c.JSON(http.StatusOK, gin.H{"status": "success", "message": "服务器信息已接收，等待管理员确认"})
	}
}

func ListServersHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		status := c.Query("status")
		servers, err := database.ListServers(db, status)
		if err != nil {
			auditEvent(c, "list_servers", status, "failure")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
			return
		}
		auditEvent(c, "list_servers", status, "success")
		c.JSON(http.StatusOK, servers)
	}
}

func GetServerHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		serial := c.Param("serial")
		s, err := database.GetServerBySerial(db, serial)
		if err != nil {
			auditEvent(c, "get_server", serial, "failure")
			c.JSON(http.StatusNotFound, gin.H{"error": "未找到服务器"})
			return
		}
		auditEvent(c, "get_server", serial, "success")
		c.JSON(http.StatusOK, s)
	}
}

func ConfirmServerHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		serial := c.Param("serial")
		if err := database.MarkServerConfirmed(db, serial); err != nil {
			auditEvent(c, "confirm_server", serial, "failure")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "确认失败"})
			return
		}
		auditEvent(c, "confirm_server", serial, "success")
		c.JSON(http.StatusOK, gin.H{"status": "success", "message": "服务器信息已确认"})
	}
}

func MarkInstalledHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		serial := c.Param("serial")
		if err := database.MarkServerInstalled(db, serial); err != nil {
			auditEvent(c, "mark_installed", serial, "failure")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "标记失败"})
			return
		}
		auditEvent(c, "mark_installed", serial, "success")
		c.JSON(http.StatusOK, gin.H{"status": "success", "message": "已标记为已安装"})
	}
}

func ListConfigsHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		configs, err := database.ListConfigs(db)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
			return
		}
		c.JSON(http.StatusOK, configs)
	}
}

func GetConfigHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, _ := strconv.Atoi(idStr)
		cfg, err := database.GetConfig(db, id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "未找到配置"})
			return
		}
		c.JSON(http.StatusOK, cfg)
	}
}

type CreateConfigRequest struct {
	Name          string `json:"name"`
	Description   string `json:"description"`
	SystemType    string `json:"systemType"`
	SystemVersion string `json:"systemVersion"`
	ConfigContent string `json:"configContent"`
	KernelParams  string `json:"kernelParams"`
	Packages      string `json:"packages"`
}

func CreateConfigHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req CreateConfigRequest
		if err := c.BindJSON(&req); err != nil {
			auditEvent(c, "create_config", "", "failure")
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求数据"})
			return
		}
		ct := &database.ConfigTemplate{
			Name:          req.Name,
			Description:   req.Description,
			SystemType:    req.SystemType,
			SystemVersion: req.SystemVersion,
			ConfigContent: req.ConfigContent,
			KernelParams:  req.KernelParams,
			Packages:      req.Packages,
			Status:        "active",
		}
		id, err := database.CreateConfig(db, ct)
		if err != nil {
			auditEvent(c, "create_config", ct.Name, "failure")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "创建失败"})
			return
		}
		auditEvent(c, "create_config", strconv.Itoa(id), "success")
		c.JSON(http.StatusOK, gin.H{"id": id})
	}
}

func UpdateConfigHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, _ := strconv.Atoi(idStr)
		var req CreateConfigRequest
		if err := c.BindJSON(&req); err != nil {
			auditEvent(c, "update_config", idStr, "failure")
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求数据"})
			return
		}
		ct := &database.ConfigTemplate{
			Name:          req.Name,
			Description:   req.Description,
			SystemType:    req.SystemType,
			SystemVersion: req.SystemVersion,
			ConfigContent: req.ConfigContent,
			KernelParams:  req.KernelParams,
			Packages:      req.Packages,
			Status:        "active",
		}
		if err := database.UpdateConfig(db, id, ct); err != nil {
			auditEvent(c, "update_config", idStr, "failure")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败"})
			return
		}
		auditEvent(c, "update_config", idStr, "success")
		c.Status(http.StatusOK)
	}
}

func ApplyConfigHandler(db *sql.DB, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, _ := strconv.Atoi(idStr)
		serial := c.Query("serial")
		if serial == "" {
			auditEvent(c, "apply_config", idStr, "failure")
			c.JSON(http.StatusBadRequest, gin.H{"error": "缺少 serial"})
			return
		}
		// 获取服务器与配置
		server, err := database.GetServerBySerial(db, serial)
		if err != nil {
			auditEvent(c, "apply_config", serial, "failure")
			c.JSON(http.StatusNotFound, gin.H{"error": "未找到服务器"})
			return
		}
		conf, err := database.GetConfig(db, id)
		if err != nil {
			auditEvent(c, "apply_config", idStr, "failure")
			c.JSON(http.StatusNotFound, gin.H{"error": "未找到配置"})
			return
		}
		g := &pxe.Generator{TFTPRoot: cfg.TFTP.Root, PXEConfig: pxe.DefaultPXEConfig()}
		if cfg.TFTP.EnableUEFI {
			g.PXEConfig.EnableUEFI = true
		}
		if err := g.GenerateConfig(server, conf); err != nil {
			auditEvent(c, "apply_config", serial, "failure")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "生成PXE配置失败"})
			return
		}
		auditEvent(c, "apply_config", serial, "success")
		c.JSON(http.StatusOK, gin.H{"status": "success", "message": "配置已应用到服务器"})
	}
}
