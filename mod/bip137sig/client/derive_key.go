package bip137sig

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/query"
	"github.com/cryptopunkscc/astrald/mod/bip137sig"
	"github.com/cryptopunkscc/astrald/mod/crypto"
)

func (client *Client) DeriveKey(ctx *astral.Context, path string, seed *bip137sig.Seed) (privateKey *crypto.PrivateKey, err error) {
	ch, err := client.queryCh(ctx, bip137sig.MethodDeriveKey, query.Args{"path": path})
	if err != nil {
		return
	}
	defer ch.Close()
	if err = ch.Send(seed); err != nil {
		return
	}
	err = ch.Switch(channel.Expect(&privateKey), channel.PassErrors)
	return
}
