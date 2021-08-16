package ip

import (
	"github.com/cryptopunkscc/astrald/node/net"
	go_net "net"
)

func BroadcastAddr(addr net.Addr) (net.Addr, error) {
	ip, ipnet, err := go_net.ParseCIDR(addr.String())
	if err != nil {
		return nil, err
	}

	if len(ipnet.Mask) == go_net.IPv4len {
		ip = ip[12:]
	}

	broadIP := make(go_net.IP, len(ipnet.Mask))

	for i := 0; i < len(ipnet.Mask); i++ {
		broadIP[i] = ip[i] | ^ipnet.Mask[i]
	}

	return net.MakeAddr(addr.Network(), broadIP.String()), nil
}
