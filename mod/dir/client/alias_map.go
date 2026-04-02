package dir

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/mod/dir"
)

func AliasMap(ctx *astral.Context) (*dir.AliasMap, error) {
	return Default().AliasMap(ctx)
}

func (client *Client) AliasMap(ctx *astral.Context) (am *dir.AliasMap, err error) {
	// query
	ch, err := client.queryCh(ctx, dir.MethodAliasMap, nil)
	if err != nil {
		return nil, err
	}

	// response
	err = ch.Switch(channel.Expect(&am), channel.PassErrors)
	return
}
