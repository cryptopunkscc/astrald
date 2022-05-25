package ip

import (
	"net"
)

func BroadcastAddr(addr net.Addr) (net.IP, error) {
	ip, ipnet, err := net.ParseCIDR(addr.String())
	if err != nil {
		return nil, err
	}

	if len(ipnet.Mask) == net.IPv4len {
		ip = ip[12:]
	}

	broadIP := make(net.IP, len(ipnet.Mask))

	for i := 0; i < len(ipnet.Mask); i++ {
		broadIP[i] = ip[i] | ^ipnet.Mask[i]
	}

	return broadIP, nil
}
