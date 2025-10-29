package pxe

import (
	"fmt"
	"pxe-manager/database"
)

// 简化的 Preseed 生成（后续可替换为模板渲染）
func GeneratePreseed(s *database.Server, t *database.ConfigTemplate) string {
	return fmt.Sprintf(`# Preseed for %s (%s)

d-i debian-installer/locale string en_US
d-i keyboard-configuration/xkb-keymap select us

# Network
# This is a placeholder; real implementation should render static config

# Custom content from template
%s

# Kernel parameters
# %s

# Packages
# %s
`, s.Serial, s.MACAddress, t.ConfigContent, t.KernelParams, t.Packages)
}
