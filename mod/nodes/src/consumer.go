package nodes

import (
	"context"
	"encoding/binary"
	"fmt"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/core"
	"github.com/cryptopunkscc/astrald/id"
)

type Consumer struct {
	*Module
	ProviderID id.Identity
}

func NewConsumer(module *Module, providerID id.Identity) *Consumer {
	return &Consumer{Module: module, ProviderID: providerID}
}

func (mod *Consumer) Relay(ctx context.Context, nonce astral.Nonce, caller id.Identity, target id.Identity) (err error) {
	params := core.Params{}
	params.SetNonce(pNonce, nonce)
	if !caller.IsZero() {
		params.SetIdentity(pSetCaller, caller)
	}
	if !target.IsZero() && !target.IsEqual(mod.ProviderID) {
		params.SetIdentity(pSetTarget, target)
	}
	q := &astral.Query{
		Nonce:  astral.NewNonce(),
		Caller: mod.node.Identity(),
		Target: mod.ProviderID,
		Query:  core.Query(mRelay, params),
	}

	conn, err := astral.Route(ctx, mod.node, q)
	if err != nil {
		return
	}
	defer conn.Close()

	var code uint8
	err = binary.Read(conn, binary.BigEndian, &code)
	if err != nil {
		return
	}
	if code != 0 {
		return fmt.Errorf("error code %d", code)
	}

	return nil
}
