package user

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/nearby"
)

var _ nearby.Composer = &Module{}

func (mod *Module) ComposeStatus(a nearby.Composition) {
	switch mod.Nearby.Mode() {
	case nearby.ModeSilent:
		// no-op
	case nearby.ModeVisible:
		if c := mod.ActiveContract(); c != nil {
			a.Attach(c)
		} else {
			alias, _ := mod.Dir.GetAlias(mod.node.Identity())
			a.Attach(nearby.NewFlag("claimable"))
			a.Attach(&nearby.PublicProfile{
				NodeID:    mod.node.Identity(),
				NodeAlias: alias,
			})
		}

	case nearby.ModeStealth:
		ac := mod.ActiveContract()
		if ac == nil {
			return
		}
		nonce := astral.NewNonce()
		a.Attach(&nearby.StealthHint{
			Commitment: nearby.ComputeCommitment(ac.Issuer, nonce),
			MaskedID:   nearby.MaskIdentity(mod.node.Identity(), ac.Issuer),
			Nonce:      nonce,
		})
	}
}
