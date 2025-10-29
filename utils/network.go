package utils

import (
	"net"
	"strings"
)

// IsIPInWhitelist 支持单 IP 与 CIDR
func IsIPInWhitelist(ip string, whitelist []string) bool {
	clientIP := net.ParseIP(ip)
	if clientIP == nil {
		return false
	}
	for _, item := range whitelist {
		if strings.Contains(item, "/") {
			_, cidrNet, err := net.ParseCIDR(item)
			if err == nil && cidrNet.Contains(clientIP) {
				return true
			}
		} else {
			if item == ip {
				return true
			}
		}
	}
	return false
}
