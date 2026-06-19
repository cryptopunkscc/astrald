package user

import (
	"errors"
	"time"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/mod/secp256k1"
	"github.com/cryptopunkscc/astrald/mod/user"
)

// Expel permanently bans nodeID from the active swarm. Only the swarm's user (the
// active contract's issuer) may expel, and the ban cannot be undone. The signed
// ban is stored, pushed to the local swarm, and the node is disconnected.
//
// Expelling suppresses swarm membership only; it does not revoke the node's
// underlying SwarmAccess contract.
func (mod *Module) Expel(ctx *astral.Context, nodeID *astral.Identity) (*user.SignedExpulsion, error) {
	ac := mod.ActiveContract()
	if ac == nil {
		return nil, user.ErrNoActiveContract
	}

	if nodeID.IsEqual(ac.Issuer) {
		return nil, errors.New("cannot expel the swarm user")
	}

	expulsion := &user.Expulsion{
		Issuer:     ac.Issuer,
		Subject:    nodeID,
		ExpelledAt: astral.Time(time.Now()),
	}

	sig, err := mod.Crypto.Sign(ctx, secp256k1.FromIdentity(ac.Issuer), expulsion)
	if err != nil {
		return nil, err
	}

	signed := &user.SignedExpulsion{Expulsion: expulsion, IssuerSig: sig}

	err = mod.db.StoreExpulsion(signed)
	if err != nil {
		return nil, err
	}

	go mod.PushToLocalSwarm(mod.ctx, signed)
	mod.applyExpulsion(nodeID)

	mod.log.Info("expelled %v from swarm", nodeID)
	return signed, nil
}

// receiveExpulsion handles an inbound signed ban. It is accepted only when issued
// by the active swarm user and carrying a valid issuer signature, then stored and
// enforced by disconnecting the subject.
func (mod *Module) receiveExpulsion(signed *user.SignedExpulsion) error {
	if signed.IsNil() {
		return objects.ErrPushRejected
	}

	ac := mod.ActiveContract()
	if ac == nil || !signed.Issuer.IsEqual(ac.Issuer) {
		return objects.ErrPushRejected
	}

	if err := mod.verifyExpulsion(signed); err != nil {
		mod.log.Errorv(1, "invalid expulsion: %v", err)
		return objects.ErrPushRejected
	}

	if err := mod.db.StoreExpulsion(signed); err != nil {
		mod.log.Errorv(1, "storing expulsion failed: %v", err)
		return objects.ErrPushRejected
	}

	mod.applyExpulsion(signed.Subject)
	return nil
}

// applyExpulsion tears down all live ties to id: it disconnects open links and
// stops the sibling link-maintenance task. Membership filtering already prevents
// re-linking once the ban is stored.
func (mod *Module) applyExpulsion(id *astral.Identity) {
	if err := mod.Nodes.CloseLinks(id); err != nil {
		mod.log.Errorv(1, "closing links to expelled %v: %v", id, err)
	}
	mod.removeSibling(id)
}

// expelledSet returns the subjects banned by issuer, keyed by identity string for
// O(1) membership filtering.
func (mod *Module) expelledSet(issuer *astral.Identity) map[string]struct{} {
	subjects, err := mod.db.ExpelledSubjects(issuer)
	if err != nil {
		mod.log.Errorv(1, "error reading expulsions: %v", err)
		return nil
	}

	set := make(map[string]struct{}, len(subjects))
	for _, s := range subjects {
		set[s.String()] = struct{}{}
	}
	return set
}

// verifyExpulsion checks that the issuer signature is present and valid.
func (mod *Module) verifyExpulsion(signed *user.SignedExpulsion) error {
	if signed.IssuerSig == nil {
		return errors.New("missing issuer signature")
	}

	return mod.Crypto.Verify(secp256k1.FromIdentity(signed.Issuer), signed.IssuerSig, signed.Expulsion)
}
