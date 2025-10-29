package pxe

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"pxe-manager/database"
)

type PXEConfig struct {
	EnableUEFI     bool
	UEFIBootPath   string
	LegacyBootPath string
	UEFIBootFile   string
	UEFIConfigFile string
	LegacyBootFile string
}

func DefaultPXEConfig() *PXEConfig {
	return &PXEConfig{
		EnableUEFI:     false,
		UEFIBootPath:   "efi/boot/",
		LegacyBootPath: "pxelinux.cfg/",
		UEFIBootFile:   "grubx64.efi",
		UEFIConfigFile: "grub.cfg",
		LegacyBootFile: "pxelinux.0",
	}
}

type Generator struct {
	TFTPRoot  string
	PXEConfig *PXEConfig
}

func (g *Generator) GenerateConfig(server *database.Server, template *database.ConfigTemplate) error {
	pxeFileName, err := FormatMACForPXE(server.MACAddress)
	if err != nil { return err }

	var content string
	switch strings.ToLower(template.SystemType) {
	case "centos", "rhel":
		content = GenerateKickstart(server, template)
	case "ubuntu", "debian":
		content = GeneratePreseed(server, template)
	default:
		return fmt.Errorf("不支持的系统类型: %s", template.SystemType)
	}

	path := filepath.Join(g.TFTPRoot, g.PXEConfig.LegacyBootPath, pxeFileName)
	if g.PXEConfig.EnableUEFI {
		// 可根据需要为 UEFI 写入额外配置文件
		uefiPath := filepath.Join(g.TFTPRoot, g.PXEConfig.UEFIBootPath, g.PXEConfig.UEFIConfigFile)
		if err := os.MkdirAll(filepath.Dir(uefiPath), 0755); err != nil {
			return fmt.Errorf("创建UEFI配置目录失败: %w", err)
		}
		if err := os.WriteFile(uefiPath, []byte(content), 0644); err != nil {
			return fmt.Errorf("写入UEFI配置失败: %w", err)
		}
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("创建PXE配置目录失败: %w", err)
	}
	return os.WriteFile(path, []byte(content), 0644)
}

func FormatMACForPXE(mac string) (string, error) {
	cleaned := strings.ReplaceAll(strings.ToLower(mac), ":", "")
	if len(cleaned) != 12 {
		return "", fmt.Errorf("invalid MAC address: %s", mac)
	}
	formatted := "01-"
	for i := 0; i < 12; i += 2 {
		if i > 0 { formatted += "-" }
		formatted += cleaned[i : i+2]
	}
	return formatted, nil
}
