package database

import (
	"database/sql"
)

// SaveServer 使用 UPSERT 以 serial 唯一进行插入或更新
func SaveServer(db *sql.DB, s *Server) error {
	_, err := db.Exec(`
	INSERT INTO servers (
		serial, hostname, ip_address, mac_address, gateway, install_time,
		sda_size, part, system_version, kernel_version, cpu_model, cpu_processor,
		mem_total, memory_num, lan_nic, lan_nic_speed, wan_nic, wan_nic_speed,
		bond_nic, bond_nic_speed, status
	) VALUES (
		?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?
	) ON CONFLICT(serial) DO UPDATE SET
		hostname=excluded.hostname,
		ip_address=excluded.ip_address,
		mac_address=excluded.mac_address,
		gateway=excluded.gateway,
		install_time=excluded.install_time,
		sda_size=excluded.sda_size,
		part=excluded.part,
		system_version=excluded.system_version,
		kernel_version=excluded.kernel_version,
		cpu_model=excluded.cpu_model,
		cpu_processor=excluded.cpu_processor,
		mem_total=excluded.mem_total,
		memory_num=excluded.memory_num,
		lan_nic=excluded.lan_nic,
		lan_nic_speed=excluded.lan_nic_speed,
		wan_nic=excluded.wan_nic,
		wan_nic_speed=excluded.wan_nic_speed,
		bond_nic=excluded.bond_nic,
		bond_nic_speed=excluded.bond_nic_speed,
		status=excluded.status;
	`,
		s.Serial, s.Hostname, s.IPAddress, s.MACAddress, s.Gateway, s.InstallTime,
		s.SdaSize, s.Part, s.SystemVersion, s.KernelVersion, s.CPUModel, s.CPUProcessor,
		s.MemTotal, s.MemoryNum, s.LanNic, s.LanNicSpeed, s.WanNic, s.WanNicSpeed,
		s.BondNic, s.BondNicSpeed, s.Status,
	)
	return err
}

// MarkServerConfirmed 将状态置为 confirmed
func MarkServerConfirmed(db *sql.DB, serial string) error {
	_, err := db.Exec(`UPDATE servers SET status='confirmed' WHERE serial=?`, serial)
	return err
}

// MarkServerInstalled 将状态置为 installed
func MarkServerInstalled(db *sql.DB, serial string) error {
	_, err := db.Exec(`UPDATE servers SET status='installed' WHERE serial=?`, serial)
	return err
}

// ListServers 根据状态筛选
func ListServers(db *sql.DB, status string) ([]Server, error) {
	q := `SELECT id, serial, hostname, ip_address, mac_address, gateway, install_time,
	      sda_size, part, system_version, kernel_version, cpu_model, cpu_processor,
	      mem_total, memory_num, lan_nic, lan_nic_speed, wan_nic, wan_nic_speed,
	      bond_nic, bond_nic_speed, status, created_at, updated_at FROM servers`
	var rows *sql.Rows
	var err error
	if status != "" {
		q += ` WHERE status=?`
		rows, err = db.Query(q, status)
	} else {
		rows, err = db.Query(q)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	res := []Server{}
	for rows.Next() {
		var s Server
		if err := rows.Scan(
			&s.ID, &s.Serial, &s.Hostname, &s.IPAddress, &s.MACAddress, &s.Gateway, &s.InstallTime,
			&s.SdaSize, &s.Part, &s.SystemVersion, &s.KernelVersion, &s.CPUModel, &s.CPUProcessor,
			&s.MemTotal, &s.MemoryNum, &s.LanNic, &s.LanNicSpeed, &s.WanNic, &s.WanNicSpeed,
			&s.BondNic, &s.BondNicSpeed, &s.Status, &s.CreatedAt, &s.UpdatedAt,
		); err != nil {
			return nil, err
		}
		res = append(res, s)
	}
	return res, rows.Err()
}

// GetServerBySerial 获取详情
func GetServerBySerial(db *sql.DB, serial string) (*Server, error) {
	row := db.QueryRow(`SELECT id, serial, hostname, ip_address, mac_address, gateway, install_time,
	      sda_size, part, system_version, kernel_version, cpu_model, cpu_processor,
	      mem_total, memory_num, lan_nic, lan_nic_speed, wan_nic, wan_nic_speed,
	      bond_nic, bond_nic_speed, status, created_at, updated_at FROM servers WHERE serial=?`, serial)
	var s Server
	if err := row.Scan(
		&s.ID, &s.Serial, &s.Hostname, &s.IPAddress, &s.MACAddress, &s.Gateway, &s.InstallTime,
		&s.SdaSize, &s.Part, &s.SystemVersion, &s.KernelVersion, &s.CPUModel, &s.CPUProcessor,
		&s.MemTotal, &s.MemoryNum, &s.LanNic, &s.LanNicSpeed, &s.WanNic, &s.WanNicSpeed,
		&s.BondNic, &s.BondNicSpeed, &s.Status, &s.CreatedAt, &s.UpdatedAt,
	); err != nil {
		return nil, err
	}
	return &s, nil
}

// Config 模板 CRUD（简单实现）
func ListConfigs(db *sql.DB) ([]ConfigTemplate, error) {
	rows, err := db.Query(`SELECT id, name, description, system_type, system_version, config_content, kernel_params, packages, status, created_at FROM config_templates`)
	if err != nil { return nil, err }
	defer rows.Close()
	var res []ConfigTemplate
	for rows.Next() {
		var c ConfigTemplate
		if err := rows.Scan(&c.ID, &c.Name, &c.Description, &c.SystemType, &c.SystemVersion, &c.ConfigContent, &c.KernelParams, &c.Packages, &c.Status, &c.CreatedAt); err != nil { return nil, err }
		res = append(res, c)
	}
	return res, rows.Err()
}

func GetConfig(db *sql.DB, id int) (*ConfigTemplate, error) {
	row := db.QueryRow(`SELECT id, name, description, system_type, system_version, config_content, kernel_params, packages, status, created_at FROM config_templates WHERE id=?`, id)
	var c ConfigTemplate
	if err := row.Scan(&c.ID, &c.Name, &c.Description, &c.SystemType, &c.SystemVersion, &c.ConfigContent, &c.KernelParams, &c.Packages, &c.Status, &c.CreatedAt); err != nil { return nil, err }
	return &c, nil
}

func CreateConfig(db *sql.DB, c *ConfigTemplate) (int64, error) {
	res, err := db.Exec(`INSERT INTO config_templates(name, description, system_type, system_version, config_content, kernel_params, packages, status) VALUES (?,?,?,?,?,?,?,?)`,
		c.Name, c.Description, c.SystemType, c.SystemVersion, c.ConfigContent, c.KernelParams, c.Packages, c.Status,
	)
	if err != nil { return 0, err }
	return res.LastInsertId()
}

func UpdateConfig(db *sql.DB, id int, c *ConfigTemplate) error {
	_, err := db.Exec(`UPDATE config_templates SET name=?, description=?, system_type=?, system_version=?, config_content=?, kernel_params=?, packages=?, status=? WHERE id=?`,
		c.Name, c.Description, c.SystemType, c.SystemVersion, c.ConfigContent, c.KernelParams, c.Packages, c.Status, id,
	)
	return err
}
