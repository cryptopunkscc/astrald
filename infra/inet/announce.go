package inet

import (
	"bytes"
	"context"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/infra"
	"github.com/cryptopunkscc/astrald/infra/ip"
	"log"
	"net"
	"time"
)

var _ infra.Announcer = &Inet{}

const announceInterval = 1 * time.Minute

func (inet *Inet) Announce(ctx context.Context) error {
	if err := inet.broadcastPresence(&presence{
		Identity: inet.localID,
		Port:     inet.getListenPort(),
		Flags:    flagDiscover,
	}); err != nil {
		return err
	}

	log.Println("[inet] announcing presence")

	go func() {
		for {
			select {
			case <-time.After(announceInterval):
				if err := inet.broadcastPresence(&presence{
					Identity: inet.localID,
					Port:     inet.getListenPort(),
					Flags:    flagNone,
				}); err != nil {
					log.Println("[inet] announce error:", err)
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	return nil
}

func (inet *Inet) broadcastPresence(p *presence) error {
	// check presence socket
	if err := inet.setupPresenceConn(); err != nil {
		return err
	}

	// prepare packet data
	var packet = &bytes.Buffer{}
	if err := cslq.Encode(packet, "v", p); err != nil {
		return err
	}

	ifaces, err := net.Interfaces()
	if err != nil {
		return err
	}

	for _, iface := range ifaces {
		if !inet.isInterfaceEnabled(iface) {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			return err
		}

		for _, addr := range addrs {
			broadcastIP, err := ip.BroadcastAddr(addr)
			if err != nil {
				return err
			}

			if ip.IsLinkLocal(broadcastIP) {
				continue
			}

			var broadcastAddr = net.UDPAddr{
				IP:   broadcastIP,
				Port: defaultPresencePort,
			}

			_, err = inet.presenceConn.WriteTo(packet.Bytes(), &broadcastAddr)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (inet *Inet) sendPresence(destAddr *net.UDPAddr, p presence) (err error) {
	// check presence socket
	if err = inet.setupPresenceConn(); err != nil {
		return
	}

	// prepare packet data
	var packet = &bytes.Buffer{}
	if err = cslq.Encode(packet, "v", p); err != nil {
		return
	}

	// send message
	_, err = inet.presenceConn.WriteTo(packet.Bytes(), destAddr)
	return
}

func (inet *Inet) isInterfaceEnabled(iface net.Interface) bool {
	return (iface.Flags&net.FlagUp != 0) &&
		(iface.Flags&net.FlagBroadcast != 0) &&
		(iface.Flags&net.FlagLoopback == 0)
}
