package info

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/astral/link"
	"github.com/cryptopunkscc/astrald/auth/id"
	_node "github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/node/contacts"
	"github.com/cryptopunkscc/astrald/node/presence"
	"io"
	"log"
)

const serviceHandle = "info"
const ModuleName = "info"

type Info struct{}

var seen map[string]struct{}

func (Info) Run(ctx context.Context, node *_node.Node) error {
	seen = map[string]struct{}{}

	port, err := node.Ports.Register(serviceHandle)
	if err != nil {
		return err
	}
	defer port.Close()

	go func() {
		for query := range port.Queries() {
			conn := query.Accept()
			conn.Write(node.Info(false).Pack())
			conn.Close()
		}
	}()

	go func() {
		for e := range node.Follow(ctx) {
			switch event := e.(type) {
			case _node.Event:
				if event.Event() == _node.EventPeerLinked {
					refreshContact(ctx, node, event.Peer.Identity())
				}

			case presence.Event:
				if event.Event() == presence.EventIdentityPresent {
					refreshContact(ctx, node, event.Identity())
				}
			}
		}
	}()

	<-ctx.Done()
	return nil
}

func (Info) String() string {
	return ModuleName
}

func refreshContact(ctx context.Context, node *_node.Node, identity id.Identity) {
	if _, found := seen[identity.PublicKeyHex()]; found {
		return
	}

	info, err := queryContact(ctx, node, identity)

	if err != nil {
		if !errors.Is(err, link.ErrRejected) {
			log.Printf("[%s] error updating info: %v\n", node.Contacts.DisplayName(identity), err)
		}
		return
	}

	seen[identity.PublicKeyHex()] = struct{}{}
	node.Contacts.AddInfo(info)

	log.Printf("[%s] updated info\n", node.Contacts.DisplayName(identity))
}

func queryContact(ctx context.Context, node *_node.Node, identity id.Identity) (*contacts.Contact, error) {
	// update peer info
	conn, err := node.Query(ctx, identity, serviceHandle)

	if err != nil {
		return nil, err
	}
	packed, err := io.ReadAll(conn)
	if err != nil {
		return nil, err
	}

	info, err := contacts.Unpack(packed)
	if err != nil {
		return nil, err
	}

	return info, nil
}
