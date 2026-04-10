package user

import (
	"slices"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/apphost"
	"github.com/cryptopunkscc/astrald/mod/auth"
	"github.com/cryptopunkscc/astrald/mod/events"
	"github.com/cryptopunkscc/astrald/mod/nodes"
	"github.com/cryptopunkscc/astrald/mod/objects"
)

var _ objects.Receiver = &Module{}

func (mod *Module) ReceiveObject(drop objects.Drop) (err error) {
	switch o := drop.Object().(type) {
	case *auth.SignedContract:
		err = mod.receiveSignedContract(drop.SenderID(), o)
		if err == nil {
			drop.Accept(true)
		}

		go func() {
			err = mod.Nodes.UpdateNodeEndpoints(mod.ctx, drop.SenderID(), o.Subject)
			if err != nil {
				mod.log.Error("syncEndpoints: %v", err)
			}
		}()

	case *apphost.SignedAppContract:
		if !slices.ContainsFunc(mod.LocalSwarm(), o.HostID.IsEqual) {
			break
		}

		drop.Accept(true)

	case *apphost.EventNewAppContract:
		switch {
		case !drop.SenderID().IsEqual(mod.node.Identity()):
			break
		case !o.Contract.HostID.IsEqual(mod.node.Identity()):
			break
		}

		mod.pushToLinkedSibs(o.Contract)

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

func (mod *Module) receiveSignedContract(s *astral.Identity, c *auth.SignedContract) error {
	isSigner := s.IsEqual(c.Subject)
	isSelf := s.IsEqual(mod.node.Identity())
	isLocalSwarmMember := slices.ContainsFunc(mod.LocalSwarm(), s.IsEqual)

	if !(isSigner || isSelf || isLocalSwarmMember) {
		return objects.ErrPushRejected
	}

	if err := mod.Auth.VerifyContract(c); err != nil {
		mod.log.Errorv(1, "invalid signed contract: %v", err)
		return objects.ErrPushRejected
	}

	if err := mod.StoreContract(c); err != nil {
		mod.log.Errorv(1, "store signed contract: %v", err)
		return objects.ErrPushRejected
	}

	go func() {
		err := mod.Nodes.UpdateNodeEndpoints(mod.ctx, s, c.Subject)
		if err != nil {
			mod.log.Error("updatingNodeEndpoint failed: %v", err)
		}
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
