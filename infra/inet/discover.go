package inet

import (
	"bytes"
	"context"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/infra"
)

var _ infra.Discoverer = &Inet{}

func (inet *Inet) Discover(ctx context.Context) (<-chan infra.Presence, error) {
	// check presence socket
	if err := inet.setupPresenceConn(); err != nil {
		return nil, err
	}

	go func() {
		<-ctx.Done()
		_ = inet.presenceConn.Close()
	}()

	outCh := make(chan infra.Presence)

	go func() {
		defer close(outCh)
		buf := make([]byte, 1024)

		for {
			n, srcAddr, err := inet.presenceConn.ReadFromUDP(buf)
			if err != nil {
				return
			}

			var p presence
			var r = bytes.NewReader(buf[:n])

			if err := cslq.Decode(r, "v", &p); err != nil {
				continue
			}

			if p.Identity.IsEqual(inet.localID) {
				// ignore our own presence
				continue
			}

			outCh <- infra.Presence{
				Identity: p.Identity,
				Addr: Addr{
					ver:  0,
					ip:   srcAddr.IP,
					port: uint16(p.Port),
				},
				Present: p.Flags&flagBye == 0,
			}

			if p.Flags&flagDiscover != 0 {
				inet.sendPresence(srcAddr, presence{
					Identity: inet.localID,
					Port:     inet.listenPort,
					Flags:    flagNone,
				})
			}
		}
	}()

	return outCh, nil
}
