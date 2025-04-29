package apphost

import (
	"fmt"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/lib/query"
	"github.com/cryptopunkscc/astrald/mod/dir"
)

var _ dir.Resolver = &Client{}

func (c *Client) ResolveIdentity(name string) (*astral.Identity, error) {
	// try to parse the public key first
	if id, err := astral.IdentityFromString(name); err == nil {
		return id, nil
	}

	// then try using node's resolver
	s, err := c.Session()
	if err != nil {
		return nil, err
	}

	params, _ := query.Marshal(map[string]string{"name": name})

	conn, err := s.Query(c.GuestID, c.HostID, "dir.resolve?"+params)
	if err != nil {
		return nil, err
	}

	ch := astral.NewChannel(conn)
	defer ch.Close()

	obj, err := ch.Read()
	if err != nil {
		return nil, err
	}

	id, ok := obj.(*astral.Identity)
	if !ok {
		return nil, fmt.Errorf("unexpected type: %s", obj.ObjectType())
	}

	return id, nil
}
