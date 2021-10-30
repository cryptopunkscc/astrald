package inet

import (
	"context"
	"github.com/cryptopunkscc/astrald/infra"
	"net"
	"strconv"
)

func (inet Inet) Discover(ctx context.Context) (<-chan infra.Presence, error) {
	outCh := make(chan infra.Presence)

	udpAddr, err := net.ResolveUDPAddr("udp", ":"+strconv.Itoa(presencePort))
	if err != nil {
		return nil, err
	}

	listener, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return nil, err
	}

	go func() {
		defer close(outCh)
		buf := make([]byte, 1024)

		for {
			n, srcAddr, err := listener.ReadFromUDP(buf)
			if err != nil {
				return
			}
			if n != presencePayloadLen {
				// Ignore packets of invalid length
				continue
			}

			peerID, port, _, err := parsePresence(buf[:n])
			if err != nil {
				return
			}

			addr, err := Parse(srcAddr.IP.String() + ":" + strconv.Itoa(int(port)))
			if err != nil {
				continue
			}

			presence := infra.Presence{
				Identity: peerID,
				Addr:     addr,
			}

			outCh <- presence
		}
	}()

	go func() {
		<-ctx.Done()
		listener.Close()
	}()

	return outCh, nil
}
