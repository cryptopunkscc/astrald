package ip

import (
	"net"
	"strings"
)

// IsAddrLocalNetwork checks if the address belongs to a local network (LAN/loopback)
func IsAddrLocalNetwork(addr net.Addr) bool {
	var ip net.IP

	if strings.Contains(addr.String(), "/") {
		// Remove CIDR mask from the address
		ipStr, _ := SplitIPMask(addr.String())
		ip = net.ParseIP(ipStr)
	} else {
		ipStr, _, _ := net.SplitHostPort(addr.String())
		ip = net.ParseIP(ipStr)
	}

	return IsPrivateIP(ip) || ip.IsLoopback()
}
