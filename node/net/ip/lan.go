package ip

import (
	"github.com/cryptopunkscc/astrald/node/net"
	_net "net"
	"strings"
)

// IsAddrLocalNetwork checks if the address belongs to a local network (LAN/loopback)
func IsAddrLocalNetwork(addr net.Addr) bool {
	var ip _net.IP

	if strings.Contains(addr.String(), "/") {
		// Remove CIDR mask from the address
		ipStr, _ := SplitIPMask(addr.String())
		ip = _net.ParseIP(ipStr)
	} else {
		ipStr, _, _ := _net.SplitHostPort(addr.String())
		ip = _net.ParseIP(ipStr)
	}

	return IsPrivateIP(ip) || ip.IsLoopback()
}
