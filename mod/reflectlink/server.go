package reflectlink

import (
	"context"
	"encoding/json"
	"github.com/cryptopunkscc/astrald/hub"
	"time"
)

func (mod *Module) runServer(ctx context.Context) error {
	port, err := mod.node.Ports.RegisterContext(ctx, portName)
	if err != nil {
		return err
	}
	defer port.Close()

	for {
		select {
		case query, ok := <-port.Queries():
			if !ok {
				return nil
			}

			// local queries happen outside of links
			if query.IsLocal() {
				query.Reject()
				break
			}

			conn, err := query.Accept()
			if err != nil {
				break
			}

			if err := mod.handleRequest(conn); err != nil {
				log.Error("handleRequest: %s", err)
			}

		case <-ctx.Done():
			return nil
		}
	}
}

func (mod *Module) handleRequest(conn *hub.Conn) error {
	var info = mod.getLocalInfo()
	var remoteAddr = conn.Link().RemoteAddr()

	info.ReflectAddr = jsonAddr{
		Network: remoteAddr.Network(),
		Address: remoteAddr.Pack(),
	}

	bytes, err := json.Marshal(info)
	if err != nil {
		return err
	}

	conn.Write(bytes)
	conn.Close()

	return nil
}

func (mod *Module) getLocalInfo() *jsonInfo {
	var info = &jsonInfo{
		AddrList: make([]jsonAddrSpec, 0),
	}

	for _, a := range mod.node.Infra().LocalAddrs() {
		info.AddrList = append(info.AddrList, jsonAddrSpec{
			Network:   a.Network(),
			Address:   a.Pack(),
			Public:    false,
			ExpiresAt: int(time.Now().Add(time.Hour * 24 * 7 * 4).Unix()),
		})
	}

	return info
}
