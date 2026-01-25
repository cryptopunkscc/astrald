package dir

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/mod/dir"
)

func (client *Client) AliasMap(ctx *astral.Context) (am *dir.AliasMap, err error) {
	// query
	ch, err := client.queryCh(ctx, "dir.alias_map", nil)
	if err != nil {
		return nil, err
	}

	// response
	err = ch.Switch(channel.Expect(&am), channel.PassErrors)
	return
}
