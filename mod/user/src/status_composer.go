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
			a.Attach(nearby.NewFlag("claimable"))
		}

	case nearby.ModeStealth:
		ac := mod.ActiveContract()
		if ac == nil {
			return
		}
		nonce := astral.NewNonce()
		a.Attach(&nearby.StealthHint{
			Commitment: nearby.ComputeCommitment(ac.UserID, nonce),
			MaskedID:   nearby.MaskIdentity(mod.node.Identity(), ac.UserID),
			Nonce:      nonce,
		})
	}
}
