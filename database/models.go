package database

// 数据模型与 API/DB 字段映射

type Server struct {
	ID            int    `json:"id" db:"id"`
	Serial        string `json:"serial" db:"serial"`
	Hostname      string `json:"hostname" db:"hostname"`
	IPAddress     string `json:"ipAddress" db:"ip_address"`
	MACAddress    string `json:"macAddress" db:"mac_address"`
	Gateway       string `json:"gateway" db:"gateway"`
	InstallTime   string `json:"installTime" db:"install_time"`
	SdaSize       string `json:"sdaSize" db:"sda_size"`
	Part          string `json:"part" db:"part"`
	SystemVersion string `json:"systemVersion" db:"system_version"`
	KernelVersion string `json:"kernelVersion" db:"kernel_version"`
	CPUModel      string `json:"cpuModel" db:"cpu_model"`
	CPUProcessor  int    `json:"cpuProcessor" db:"cpu_processor"`
	MemTotal      int    `json:"memTotal" db:"mem_total"`
	MemoryNum     int    `json:"memoryNum" db:"memory_num"`
	LanNic        string `json:"lanNic" db:"lan_nic"`
	LanNicSpeed   string `json:"lanNicSpeed" db:"lan_nic_speed"`
	WanNic        string `json:"wanNic" db:"wan_nic"`
	WanNicSpeed   string `json:"wanNicSpeed" db:"wan_nic_speed"`
	BondNic       string `json:"bondNic" db:"bond_nic"`
	BondNicSpeed  string `json:"bondNicSpeed" db:"bond_nic_speed"`
	Status        string `json:"status" db:"status"`
	CreatedAt     string `json:"createdAt" db:"created_at"`
	UpdatedAt     string `json:"updatedAt" db:"updated_at"`
}

type ConfigTemplate struct {
	ID            int    `json:"id" db:"id"`
	Name          string `json:"name" db:"name"`
	Description   string `json:"description" db:"description"`
	SystemType    string `json:"systemType" db:"system_type"`
	SystemVersion string `json:"systemVersion" db:"system_version"`
	ConfigContent string `json:"configContent" db:"config_content"`
	KernelParams  string `json:"kernelParams" db:"kernel_params"`
	Packages      string `json:"packages" db:"packages"`
	Status        string `json:"status" db:"status"`
	CreatedAt     string `json:"createdAt" db:"created_at"`
}
