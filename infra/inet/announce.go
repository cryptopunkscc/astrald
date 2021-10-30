package inet

import (
	"bytes"
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/enc"
	"github.com/cryptopunkscc/astrald/infra/ip"
	"log"
	"net"
	"strconv"
	"strings"
	"time"
)

const presencePort = 8829
const presencePayloadLen = 36
const presenceInterval = time.Minute

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

func (inet *Inet) announceOnInterface(ctx context.Context, id id.Identity, iface *ip.Interface) error {
	for addr := range iface.WatchAddrs(ctx) {
		// Skip non-private addresses
		if !addr.IsPrivate() {
			continue
		}

		go func(addr *ip.Addr) {
			err := inet.announceOnAddress(ctx, id, addr)
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
	return nil
}

func (inet *Inet) announceOnAddress(ctx context.Context, id id.Identity, addr *ip.Addr) error {
	broadAddr, err := ip.BroadcastAddr(addr)
	if err != nil {
		return err
	}

	broadStr := net.JoinHostPort(broadAddr.String(), strconv.Itoa(presencePort))

	broadConn, err := net.Dial("udp", broadStr)
	if err != nil {
		return err
	}
	defer broadConn.Close()

	log.Println("start announce to", broadStr)
	defer log.Println("stop announce to", broadStr)

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
	enc.WriteIdentity(buf, id)
	enc.Write(buf, port)
	enc.Write(buf, flags)
	return buf.Bytes()
}

func parsePresence(data []byte) (_id id.Identity, port uint16, flags uint8, err error) {
	if len(data) != presencePayloadLen {
		return _id, port, flags, errors.New("invalid data")
	}
	r := bytes.NewReader(data)
	_id, _ = enc.ReadIdentity(r)
	port, _ = enc.ReadUint16(r)
	flags, _ = enc.ReadUint8(r)
	return
}
