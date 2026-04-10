package nearby

import (
	"bytes"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/auth"
	"github.com/cryptopunkscc/astrald/mod/nearby"
)

func (mod *Module) ResolveStatus(status *nearby.StatusMessage) *astral.Identity {
	// visible mode: identity from signed contract
	if c, ok := astral.First[*auth.SignedContract](status.Attachments.Objects()); ok {
		return c.Subject
	}

	// stealth mode: verify commitment and unmask
	if hint, ok := astral.First[*nearby.StealthHint](status.Attachments.Objects()); ok {
		userID := mod.Deps.User.Identity()
		if userID == nil {
			return nil
		}

		if !bytes.Equal(nearby.ComputeCommitment(userID, hint.Nonce), hint.Commitment) {
			return nil
		}

		id, err := nearby.UnmaskIdentity(hint.MaskedID, userID)
		if err != nil {
			return nil
		}
		return id
	}

	return nil
}
