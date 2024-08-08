package nodes

import (
	"context"
	"encoding/binary"
	"fmt"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/lib/query"
)

type Consumer struct {
	*Module
	ProviderID *astral.Identity
}

func NewConsumer(module *Module, providerID *astral.Identity) *Consumer {
	return &Consumer{Module: module, ProviderID: providerID}
}

func (mod *Consumer) Relay(ctx context.Context, nonce astral.Nonce, caller *astral.Identity, target *astral.Identity) (err error) {
	conn, err := query.Run(mod.node, target, mRelay, relayArgs{
		Nonce:     nonce,
		SetCaller: caller,
		SetTarget: target,
	})
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
