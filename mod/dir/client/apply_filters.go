package dir

import (
	"strings"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/query"
	"github.com/cryptopunkscc/astrald/mod/dir"
)

func ApplyFilters(ctx *astral.Context, identity *astral.Identity, filters ...string) (bool, error) {
	return Default().ApplyFilters(ctx, identity, filters...)
}

func (client *Client) ApplyFilters(ctx *astral.Context, identity *astral.Identity, filters ...string) (bool, error) {
	ch, err := client.queryCh(ctx, dir.MethodSetAlias, query.Args{
		"id":      identity,
		"filters": strings.Join(filters, ","),
	})
	if err != nil {
		return false, err
	}

	var match *astral.Bool
	err = ch.Switch(channel.Expect(&match), channel.PassErrors)
	if err != nil {
		return false, err
	}

	return (bool)(*match), nil
}
