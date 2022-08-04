package contacts

import (
	"context"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/node/contacts"
	proto "github.com/cryptopunkscc/astrald/proto/contacts"
	"io"
	"log"
)

type Contacts struct {
	node *node.Node
}

func (p Contacts) Run(ctx context.Context, n *node.Node) error {
	port, err := n.Ports.Register(proto.Port)
	if err != nil {
		return err
	}
	defer port.Close()

	p.node = n
	go func() {
		for query := range port.Queries() {
			conn, err := query.Accept()
			if err != nil {
				log.Println("Cannot accept query", err)
				continue
			}
			go func(conn io.ReadWriteCloser) {
				defer conn.Close()
				_ = proto.Serve(p, conn)
			}(conn)
		}
	}()

	<-ctx.Done()
	return nil
}

func (p Contacts) Contacts() <-chan *contacts.Contact {
	return p.node.Contacts.All()
}

func (p Contacts) String() string {
	return proto.Port
}
