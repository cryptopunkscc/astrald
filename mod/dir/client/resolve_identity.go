package dir

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/query"
)

func (client *Client) ResolveIdentity(ctx *astral.Context, name string) (identity *astral.Identity, err error) {
	// try to parse the public key first
	if id, err := astral.ParseIdentity(name); err == nil {
		return id, nil
	}

	// check cache
	if id, ok := client.resolveCache.Get(name); ok {
		return id, nil
	}

	// then try using host's resolver
	ch, err := client.queryCh(ctx, "dir.resolve", query.Args{"name": name})
	if err != nil {
		return nil, err
	}
	defer ch.Close()

	// fetch the identity from the channel
	err = ch.Switch(channel.Expect(&identity), channel.PassErrors)
	if err != nil {
		return nil, err
	}

	// cache results
	client.resolveCache.Set(name, identity)

	return
}

func ResolveIdentity(ctx *astral.Context, name string) (*astral.Identity, error) {
	return Default().ResolveIdentity(ctx, name)
}
