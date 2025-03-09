package user

import (
	"context"
	"fmt"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/core"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/lib/query"
	"github.com/cryptopunkscc/astrald/mod/user"
	"time"
)

type Consumer struct {
	mod        *Module
	consumerID *astral.Identity
	providerID *astral.Identity
}

func NewConsumer(mod *Module, consumerID *astral.Identity, providerID *astral.Identity) *Consumer {
	return &Consumer{
		mod:        mod,
		consumerID: consumerID,
		providerID: providerID,
	}
}

func (con *Consumer) Claim(ctx context.Context, d time.Duration) (err error) {
	if d == 0 {
		d = defaultContractValidity
	}

	var contract = &user.SignedNodeContract{
		NodeContract: &user.NodeContract{
			UserID:    con.mod.userID,
			NodeID:    con.providerID,
			ExpiresAt: astral.Time(time.Now().Add(d).UTC()),
		},
	}

	contract.UserSig, err = con.mod.Keys.SignASN1(contract.UserID, contract.Hash())
	if err != nil {
		return fmt.Errorf("sign contract: %w", err)
	}

	var q = astral.NewQuery(con.consumerID, con.providerID, core.Query(methodClaim, nil))

	conn, err := query.Route(ctx, con.mod.node, q)
	if err != nil {
		return err
	}
	defer conn.Close()

	go func() {
		<-ctx.Done()
		conn.Close()
	}()

	_, err = contract.WriteTo(conn)
	if err != nil {
		return fmt.Errorf("write contract: %w", err)
	}

	err = cslq.Decode(conn, "[s]c", &contract.NodeSig)
	if err != nil {
		return fmt.Errorf("read signature: %w", err)
	}

	err = con.mod.Keys.VerifyASN1(contract.NodeID, contract.Hash(), contract.NodeSig)
	if err != nil {
		return fmt.Errorf("received invalid signature: %w", err)
	}

	return con.mod.SaveSignedNodeContract(contract)
}
