package roam

import (
	"context"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/mux"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/node/link"
	"github.com/cryptopunkscc/astrald/query"
)

type DropService struct {
	*Module
}

func (service *DropService) Run(ctx context.Context) error {
	s, err := service.node.Services().Register(ctx, service.node.Identity(), dropServiceName, service)
	if err != nil {
		return err
	}
	<-s.Done()
	return nil
}

func (service *DropService) RouteQuery(ctx context.Context, q query.Query, remoteWriter net.SecureWriteCloser) (net.SecureWriteCloser, error) {
	if linker, ok := remoteWriter.(query.Linker); ok {
		if l, ok := linker.Link().(*link.Link); ok {
			return query.Accept(q, remoteWriter, func(conn net.SecureConn) {
				service.serve(conn, l)
			})
		}
	}

	return nil, link.ErrRejected
}

func (service *DropService) serve(client net.SecureConn, l *link.Link) {
	defer client.Close()

	var moveID, newRemotePort int

	if err := cslq.Decode(client, "cs", &moveID, &newRemotePort); err != nil {
		return
	}

	movable, found := service.moves[moveID]
	if !found {
		return
	}
	delete(service.moves, moveID)

	// allocate a new input stream and write its id
	var newReader = link.NewPortReader()
	newLocalPort, err := l.Mux().BindAny(newReader)
	if err != nil {
		return
	}

	movable.SetFallbackReader(newReader)

	cslq.Encode(client, "s", newLocalPort)

	// replace the output stream and finalize the move
	newWriter := mux.NewFrameWriter(l.Mux(), newRemotePort)
	movable.SetWriter(newWriter)
	movable.Attach(l)
}
