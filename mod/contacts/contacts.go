package contacts

import (
	"context"
	_log "github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/node/contacts"
	proto "github.com/cryptopunkscc/astrald/proto/contacts"
	"io"
)

type Contacts struct {
	node *node.Node
}

var log = _log.Tag("mod.contacts")

func (mod *Contacts) Run(ctx context.Context) error {
	port, err := mod.node.Ports.Register(proto.Port)
	if err != nil {
		return err
	}
	defer port.Close()

	go func() {
		for query := range port.Queries() {
			conn, err := query.Accept()
			if err != nil {
				log.Error("accept: %s", err)
				continue
			}
			go func(conn io.ReadWriteCloser) {
				defer conn.Close()
				_ = proto.Serve(mod, conn)
			}(conn)
		}
	}()

	<-ctx.Done()
	return nil
}

func (mod *Contacts) Contacts() <-chan *contacts.Contact {
	return mod.node.Contacts.All()
}
