package info

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/link"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/node/peers"
	"github.com/cryptopunkscc/astrald/node/presence"
	"io"
	"log"
	"time"
)

const serviceHandle = "info"

type Info struct {
	node *node.Node
}

var seen map[string]struct{}

type Addr struct {
	Network string
	Data    string
	Public  bool
}

type Node struct {
	Alias     string
	Addresses []Addr
}

func (mod *Info) Run(ctx context.Context) error {
	seen = map[string]struct{}{}

	port, err := mod.node.Ports.Register(serviceHandle)
	if err != nil {
		return err
	}
	defer port.Close()

	go func() {
		for query := range port.Queries() {
			conn, err := query.Accept()
			if err != nil {
				continue
			}

			info := getInfo(mod.node)
			bytes, _ := json.Marshal(info)

			conn.Write(bytes)
			conn.Close()
		}
	}()

	go func() {
		for e := range mod.node.Subscribe(ctx.Done()) {
			switch event := e.(type) {
			case peers.EventLinked:
				refreshContact(ctx, mod.node, event.Peer.Identity())

			case presence.EventIdentityPresent:
				mod.node.Contacts.Find(event.Identity, true).Add(event.Addr, time.Time{})
				mod.node.Contacts.Save()

				refreshContact(ctx, mod.node, event.Identity)
			}
		}
	}()

	<-ctx.Done()
	return nil
}

func getInfo(node *node.Node) *Node {
	info := &Node{
		Addresses: make([]Addr, 0),
	}

	info.Alias = node.Alias()

	for _, a := range node.Infra.Addresses() {
		info.Addresses = append(info.Addresses, Addr{
			Network: a.Network(),
			Data:    hex.EncodeToString(a.Pack()),
			Public:  false,
		})
	}

	return info
}

func refreshContact(ctx context.Context, node *node.Node, identity id.Identity) {
	if _, found := seen[identity.PublicKeyHex()]; found {
		return
	}

	info, err := queryContact(ctx, node, identity)

	if err != nil {
		if !errors.Is(err, link.ErrRejected) {
			log.Printf("(info) [%s] update error: %v\n", node.Contacts.DisplayName(identity), err)
		}
		return
	}

	seen[identity.PublicKeyHex()] = struct{}{}

	c := node.Contacts.Find(identity, true)
	if c.Alias() == "" {
		c.SetAlias(info.Alias)
	}

	for _, a := range info.Addresses {
		data, err := hex.DecodeString(a.Data)
		if err != nil {
			continue
		}

		addr, err := node.Infra.Unpack(a.Network, data)
		if err != nil {
			continue
		}

		c.Add(addr, time.Time{})
	}

	node.Contacts.Save()

	//TODO: Emit an event for logging?
	//log.Printf("(info) [%s] updated\n", node.Contacts.DisplayName(identity))
}

func queryContact(ctx context.Context, node *node.Node, identity id.Identity) (*Node, error) {
	// update peer info
	qctx, _ := context.WithTimeout(ctx, 10*time.Second)
	conn, err := node.Query(qctx, identity, serviceHandle)

	if err != nil {
		return nil, err
	}
	data, err := io.ReadAll(conn)
	if err != nil {
		return nil, err
	}

	var info = &Node{}
	if err := json.Unmarshal(data, &info); err != nil {
		return nil, err
	}

	return info, nil
}
