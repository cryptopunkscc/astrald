package inet

import (
	"encoding/binary"
	"errors"
	"github.com/cryptopunkscc/astrald/infra"
	"net"
	"strconv"
	"strings"
)

const NetworkName = "inet"

type Addr struct {
	ip   net.IP
	port uint16
}

var _ infra.Addr = Addr{}

func Parse(s string) (addr Addr, err error) {
	ipport := strings.Split(s, ":")
	if len(ipport) > 2 {
		return addr, errors.New("invalid address")
	}

	addr.ip = net.ParseIP(ipport[0])
	if addr.ip == nil {
		return addr, errors.New("invalid ip")
	}

	if len(ipport) == 2 {
		port, err := strconv.Atoi(ipport[1])

		if (err != nil) || (port < 0) || (port > 65535) {
			return addr, errors.New("invalid port")
		}

		addr.port = uint16(port)
	}

	return
}

func Unpack(addr []byte) (Addr, error) {
	if len(addr) != 6 {
		return Addr{}, errors.New("invalid data length")
	}
	ip := make([]byte, 4)
	copy(ip, addr[0:4])
	port := binary.BigEndian.Uint16(addr[4:6])
	return Addr{
		ip:   ip,
		port: port,
	}, nil
}

func (addr Addr) Network() string {
	return NetworkName
}

func (addr Addr) String() string {
	str := addr.ip.String()
	if addr.port != 0 {
		str = str + ":" + strconv.Itoa(int(addr.port))
	}
	return str
}

func (addr Addr) Pack() []byte {
	bytes := make([]byte, 6)
	copy(bytes[0:4], addr.ip[len(addr.ip)-4:])
	binary.BigEndian.PutUint16(bytes[4:6], addr.port)

	return bytes
}
