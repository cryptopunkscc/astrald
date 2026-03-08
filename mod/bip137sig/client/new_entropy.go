package bip137sig

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/query"
	"github.com/cryptopunkscc/astrald/mod/bip137sig"
)

func (client *Client) NewEntropy(ctx *astral.Context, bits int) (entropy *bip137sig.Entropy, err error) {
	ch, err := client.queryCh(ctx, "bip137sig.new_entropy", query.Args{"bits": bits})
	if err != nil {
		return
	}
	defer ch.Close()
	err = ch.Switch(channel.Expect(&entropy), channel.PassErrors)
	return
}
