package dir

import (
	"strings"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/query"
)

func (client *Client) ApplyFilters(ctx *astral.Context, identity *astral.Identity, filters ...string) (bool, error) {
	ch, err := client.queryCh(ctx, "dir.set_alias", query.Args{
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
