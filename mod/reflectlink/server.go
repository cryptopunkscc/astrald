package reflectlink

import (
	"context"
	"encoding/json"
	"github.com/cryptopunkscc/astrald/node/services"
	"time"
)

func (mod *Module) runServer(ctx context.Context) error {
	var queries = services.NewQueryChan(1)
	service, err := mod.node.Services().Register(ctx, mod.node.Identity(), portName, queries.Push)
	if err != nil {
		return err
	}

	go func() {
		<-service.Done()
		close(queries)
	}()

	for query := range queries {
		// local queries happen outside of links
		if query.Source() == services.SourceLocal {
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
	}

	return nil
}

func (mod *Module) handleRequest(conn *services.Conn) error {
	var info = mod.getLocalInfo()
	var remoteAddr = conn.Link().RemoteEndpoint()

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

	for _, a := range mod.node.Infra().Endpoints() {
		info.AddrList = append(info.AddrList, jsonAddrSpec{
			Network:   a.Network(),
			Address:   a.Pack(),
			Public:    false,
			ExpiresAt: int(time.Now().Add(time.Hour * 24 * 7 * 4).Unix()),
		})
	}

	return info
}
