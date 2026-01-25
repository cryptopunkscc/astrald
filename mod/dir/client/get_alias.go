package dir

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/query"
)

func (client *Client) GetAlias(ctx *astral.Context, identity *astral.Identity) (alias string, err error) {
	if alias, ok := client.aliasCache.Get(identity.String()); ok {
		return alias, nil
	}

	// query
	ch, err := client.queryCh(ctx, "dir.get_alias", query.Args{
		"id": identity,
	})
	if err != nil {
		return "", err
	}

	err = ch.Switch(channel.ExpectString[*astral.String8](&alias), channel.PassErrors)
	if err != nil {
		return "", err
	}

	// save to cache
	client.aliasCache.Set(identity.String(), alias)

	return
}

func GetAlias(ctx *astral.Context, identity *astral.Identity) (string, error) {
	return Default().GetAlias(ctx, identity)
}
