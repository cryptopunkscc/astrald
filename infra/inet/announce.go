package inet

import (
	"bytes"
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/infra/ip"
	"log"
	"net"
	"strconv"
	"strings"
	"time"
)

func (inet *Inet) Announce(ctx context.Context, id id.Identity) error {
	go func() {
		ifaceCh := ip.WatchInterfaces(ctx)
		for {
			select {
			case iface := <-ifaceCh:
				go inet.announceOnInterface(ctx, id, iface)
			case <-ctx.Done():
				return
			}
		}
	}()
	return nil
}

func (inet *Inet) announceOnInterface(ctx context.Context, id id.Identity, iface *ip.Interface) {
	if inet.config.AnnounceOnlyIface != "" {
		if inet.config.AnnounceOnlyIface != iface.Name() {
			return
		}
	}

	for addr := range iface.WatchAddrs(ctx) {
		// Skip non-private addresses
		if !addr.IsPrivate() {
			continue
		}

		go func(addr *ip.Addr) {
			err := inet.announceOnAddress(ctx, id, addr, iface.Name())
			if err != nil {
				// this error only means we lost the IP address, so it should be ignored
				netUnreachable := strings.Contains(err.Error(), "network is unreachable")
				ctxCanceled := errors.Is(err, context.Canceled)

				if !(netUnreachable || ctxCanceled) {
					log.Println("announce error:", err)
				}
			}
		}(addr)
	}
}

func (inet *Inet) announceOnAddress(ctx context.Context, id id.Identity, addr *ip.Addr, ifaceName string) error {
	broadAddr, err := ip.BroadcastAddr(addr)
	if err != nil {
		return err
	}

	broadStr := net.JoinHostPort(broadAddr.String(), strconv.Itoa(defaultPresencePort))

	broadConn, err := net.Dial("udp", broadStr)
	if err != nil {
		return err
	}
	defer broadConn.Close()

	log.Printf("announce: %s (%s)\n", broadStr, ifaceName)
	defer log.Printf("stop announce: %s (%s)\n", broadStr, ifaceName)

	announcement := makePresence(id, inet.listenPort, 0)

	for {
		_, err = broadConn.Write(announcement)
		if err != nil {
			return err
		}

		select {
		case <-time.After(presenceInterval):
			continue
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func makePresence(id id.Identity, port uint16, flags uint8) []byte {
	buf := &bytes.Buffer{}
	cslq.WriteIdentity(buf, id)
	cslq.Write(buf, port)
	cslq.Write(buf, flags)
	return buf.Bytes()
}

func parsePresence(data []byte) (_id id.Identity, port uint16, flags uint8, err error) {
	if len(data) != presencePayloadLen {
		return _id, port, flags, errors.New("invalid data")
	}
	r := bytes.NewReader(data)
	_id, _ = cslq.ReadIdentity(r)
	port, _ = cslq.ReadUint16(r)
	flags, _ = cslq.ReadUint8(r)
	return
}
