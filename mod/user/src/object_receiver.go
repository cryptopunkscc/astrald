package user

import (
	"slices"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/auth"
	"github.com/cryptopunkscc/astrald/mod/events"
	"github.com/cryptopunkscc/astrald/mod/nodes"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/mod/user"
)

var _ objects.Receiver = &Module{}

func (mod *Module) ReceiveObject(drop objects.Drop) (err error) {
	switch o := drop.Object().(type) {
	case *auth.SignedContract:
		err = mod.receiveSignedContract(drop.SenderID(), o)
		if err == nil {
			drop.Accept(true)
		}
	case *events.Event:
		switch e := o.Data.(type) {
		case *nodes.StreamCreatedEvent:
			if e.StreamCount == 1 && slices.ContainsFunc(mod.LocalSwarm(), e.RemoteIdentity.IsEqual) {
				go mod.pushActiveContract(e.RemoteIdentity)
				go mod.syncSiblings(e.RemoteIdentity)

				mod.Scheduler.Schedule(mod.NewSyncNodesTask(e.RemoteIdentity))
				drop.Accept(false)
			}
		}
	}

	return nil
}

func (mod *Module) receiveSignedContract(sender *astral.Identity, signed *auth.SignedContract) error {
	ac := mod.ActiveContract()

	isIssuerUser := ac != nil && signed.Issuer.IsEqual(ac.Issuer)
	isSubjectSwarmMember := slices.ContainsFunc(mod.LocalSwarm(), signed.Subject.IsEqual)
	isIssuerSwarmMember := slices.ContainsFunc(mod.LocalSwarm(), signed.Issuer.IsEqual)

	if !(isIssuerUser || isSubjectSwarmMember || isIssuerSwarmMember) {
		return objects.ErrPushRejected
	}

	if err := mod.Auth.VerifyContract(signed); err != nil {
		mod.log.Errorv(1, "invalid signed contract: %v", err)
		return objects.ErrPushRejected
	}

	err := mod.Auth.IndexContract(mod.ctx, signed)
	if err != nil {
		mod.log.Errorv(1, "indexing signed contract failed: %v", err)
		return objects.ErrPushRejected
	}

	go func() {
		if !user.IsNodeContract(signed.Contract) {
			return
		}

		err := mod.Nodes.UpdateNodeEndpoints(mod.ctx, sender, signed.Subject)
		if err != nil {
			mod.log.Error("syncEndpoints: %v", err)
		}

		mod.runSiblingLinker()
	}()

	return nil
}

func (mod *Module) pushActiveContract(remoteIdentity *astral.Identity) {
	contract := mod.ActiveContract()
	if contract == nil {
		return
	}

	mod.Objects.Push(mod.ctx, remoteIdentity, contract)
}
