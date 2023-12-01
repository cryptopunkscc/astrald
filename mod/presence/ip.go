package presence

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

func IsLinkLocal(ip net.IP) bool {
	if ip := ip.To4(); ip != nil {
		return ip[0] == 169 && ip[1] == 254
	}
	return ip[0] == 0xfe && ip[1] == 0x80
}
