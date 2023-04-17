package reflectlink

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/cryptopunkscc/astrald/node/link"
	"io"
	"time"
)

func (mod *Module) runClient(ctx context.Context) error {
	for e := range mod.node.Events().Subscribe(ctx) {
		switch event := e.(type) {
		case link.EventLinkEstablished:
			go func() {
				if err := mod.processEvent(ctx, event); err != nil {
					if !errors.Is(err, link.ErrRejected) {
						log.Error("query: %s", err)
					} else {
						log.Errorv(1, "query: %s", err)
					}
				}
			}()
		}
	}

	return nil
}

func (mod *Module) processEvent(ctx context.Context, event link.EventLinkEstablished) error {
	info, err := mod.queryLinkInfo(ctx, event.Link)
	if err != nil {
		return err
	}

	return mod.processLinkInfo(event.Link, info)
}

func (mod *Module) processLinkInfo(link *link.Link, jInfo *jsonInfo) error {
	remoteID := link.RemoteIdentity()
	info := &Info{Addrs: make([]AddrSpec, 0)}

	for _, a := range jInfo.AddrList {
		addr, err := mod.node.Infra().Unpack(a.Network, a.Address)
		if err != nil {
			continue
		}

		var expiresAt = time.Unix(int64(a.ExpiresAt), 0)

		info.Addrs = append(info.Addrs, AddrSpec{
			Addr:      addr,
			ExpiresAt: expiresAt,
			Public:    a.Public,
		})

		mod.node.Tracker().Add(remoteID, addr, expiresAt)
	}

	if a, err := mod.node.Infra().Unpack(jInfo.ReflectAddr.Network, jInfo.ReflectAddr.Address); err == nil {
		info.ReflectAddr = a
	}

	mod.node.Events().Emit(EventLinkReflected{Link: link, Info: info})

	return nil
}

func (mod *Module) queryLinkInfo(ctx context.Context, link *link.Link) (*jsonInfo, error) {
	conn, err := link.Query(ctx, portName)
	if err != nil {
		return nil, err
	}

	bytes, err := io.ReadAll(conn)
	info := &jsonInfo{}
	if err := json.Unmarshal(bytes, &info); err != nil {
		return nil, err
	}

	return info, nil
}
