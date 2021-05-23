package lan

import (
	"github.com/cryptopunkscc/astrald/node/auth/id"
	"github.com/cryptopunkscc/astrald/node/net"
	goNet "net"
	"strings"
)

type driver struct {
	port          uint16
	localIdentity id.Identity
}

var _ net.Driver = &driver{}

// NewDriver returns a new instance of LAN driver
func NewDriver(localIdentity id.Identity, port uint16) *driver {
	return &driver{
		localIdentity: localIdentity,
		port:          port,
	}
}

// Network returns the name of the network
func (drv *driver) Network() string {
	return "lan"
}

// isAddrLocalNetwork checks if the address belongs to a local network (LAN/loopback)
func isAddrLocalNetwork(addr net.Addr) bool {
	var ip goNet.IP

	if strings.Contains(addr.String(), "/") {
		// Remove CIDR mask from the address
		ipStr, _ := net.SplitIPMask(addr.String())
		ip = goNet.ParseIP(ipStr)
	} else {
		ipStr, _, _ := goNet.SplitHostPort(addr.String())
		ip = goNet.ParseIP(ipStr)
	}

	return net.IsPrivateIP(ip) || ip.IsLoopback()
}
