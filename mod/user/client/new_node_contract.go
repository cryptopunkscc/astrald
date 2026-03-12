package user

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/query"
	"github.com/cryptopunkscc/astrald/mod/user"
)

func (client *Client) NewNodeContract(ctx *astral.Context, alias string) (contract *user.NodeContract, err error) {
	ch, err := client.queryCh(ctx, "user.new_node_contract", query.Args{"user": alias})
	if err != nil {
		return
	}
	defer ch.Close()
	err = ch.Switch(channel.Expect(&contract), channel.PassErrors)
	return
}
