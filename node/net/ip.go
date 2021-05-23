package net

import "net"

// IsPrivateIP is temporary until Go's IP.IsPrivate() lands a release
func IsPrivateIP(ip net.IP) bool {
	if ip4 := ip.To4(); ip4 != nil {
		return ip4[0] == 10 ||
			(ip4[0] == 172 && ip4[1]&0xf0 == 16) ||
			(ip4[0] == 192 && ip4[1] == 168)
	}
	return len(ip) == net.IPv6len && ip[0]&0xfe == 0xfc
}
