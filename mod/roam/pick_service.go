package roam

import (
	"context"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/node/oldlink"
)

type PickService struct {
	*Module
}

func (service *PickService) Run(ctx context.Context) error {
	s, err := service.node.Services().Register(ctx, service.node.Identity(), pickServiceName, service)
	if err != nil {
		return err
	}
	<-s.Done()
	return nil
}

func (service *PickService) RouteQuery(ctx context.Context, query net.Query, caller net.SecureWriteCloser) (net.SecureWriteCloser, error) {
	if linker, ok := caller.(net.Linker); ok {
		if l, ok := linker.Link().(*oldlink.Link); ok {
			return net.Accept(query, caller, func(conn net.SecureConn) {
				service.serve(conn, l)
			})
		}
	}

	return nil, net.ErrRejected
}

func (service *PickService) serve(client net.SecureConn, l *oldlink.Link) {
	defer client.Close()

	var remotePort int

	// read remote port of the connection to pick
	cslq.Decode(client, "s", &remotePort)

	// find the connection
	target := l.Conns().FindByRemotePort(remotePort)
	if target == nil {
		return
	}

	// allocate a new move for the connection
	moveID := service.allocMove(target)
	if moveID != -1 {
		cslq.Encode(client, "c", moveID)
	}
}
