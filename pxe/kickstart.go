package pxe

import (
	"fmt"
	"pxe-manager/database"
)

// 简化的 Kickstart 生成（后续可替换为模板渲染）
func GenerateKickstart(s *database.Server, t *database.ConfigTemplate) string {
	return fmt.Sprintf(`#version=RHEL8
# Generated for %s (%s)

install
url --url="http://mirror.example.com/centos"
lang en_US.UTF-8
keyboard us
timezone Asia/Shanghai
network --hostname=%s --device=%s --bootproto=static --ip=%s --gateway=%s
rootpw --plaintext pxe-default

%s

%s

%s

reboot
`, s.Serial, s.MACAddress, s.Hostname, s.LanNic, s.IPAddress, s.Gateway, t.ConfigContent, t.KernelParams, t.Packages)
}
