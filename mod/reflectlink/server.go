package reflectlink

import (
	"context"
	"encoding/json"
	"github.com/cryptopunkscc/astrald/node/services"
	"time"
)

func (mod *Module) runServer(ctx context.Context) error {
	_, err := mod.node.Services().Register(ctx, mod.node.Identity(), serviceName, mod.handleQuery)
	if err != nil {
		return err
	}

	<-ctx.Done()

	return nil
}

func (mod *Module) handleQuery(ctx context.Context, query *services.Query) error {
	mod.log.Logv(2, "query from %s", query.RemoteIdentity())
	if query.Source() == services.SourceLocal {
		query.Reject()
		return nil
	}

	conn, err := query.Accept()
	if err != nil {
		return nil
	}

	if err := mod.sendInfo(conn); err != nil {
		mod.log.Error("sendInfo: %s", err)
	}

	return nil
}

func (mod *Module) sendInfo(conn *services.Conn) error {
	defer conn.Close()

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
