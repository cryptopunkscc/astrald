package user

import (
	"slices"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/apphost"
	"github.com/cryptopunkscc/astrald/mod/events"
	"github.com/cryptopunkscc/astrald/mod/nodes"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/mod/user"
)

var _ objects.Receiver = &Module{}

func (mod *Module) ReceiveObject(drop objects.Drop) (err error) {
	switch o := drop.Object().(type) {
	case *user.SignedNodeContract:
		err = mod.receiveSignedNodeContract(drop.SenderID(), o)
		if err == nil {
			drop.Accept(true)
		}

		// note: temporary solution where we sync endpoints
		go func() {
			err = mod.Nodes.UpdateNodeEndpoints(mod.ctx, drop.SenderID(), o.NodeID)
			if err != nil {
				mod.log.Error("syncEndpoints: %v", err)
			}
		}()

	case *user.SignedNodeContractRevocation:
		contract, err := mod.GetNodeContract(o.ContractID)
		if err != nil {
			mod.log.Errorv(1, "failed to get contract: %v", err)
			drop.Accept(false)
		}

		err = mod.receiveSignedNodeContractRevocation(drop.SenderID(), o, contract)
		if err == nil {
			drop.Accept(true)
		}

	case *apphost.AppContract:
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

	case *user.Notification:
		err = mod.onNotification(drop.SenderID(), o)
		if err == nil {
			drop.Accept(false)
		}

	case *events.Event:
		switch e := o.Data.(type) {
		case *nodes.StreamCreatedEvent:
			if e.StreamCount == 1 && slices.ContainsFunc(mod.LocalSwarm(), e.RemoteIdentity.IsEqual) {
				go mod.pushActiveContract(e.RemoteIdentity)

				mod.Scheduler.Schedule(mod.NewSyncNodesTask(e.RemoteIdentity))
				drop.Accept(false)
			}
		}
	}

	return nil
}

func (mod *Module) receiveSignedNodeContract(s *astral.Identity, c *user.SignedNodeContract) error {
	// reject contracts coming from neither the signing node nor local node
	// note: temporarily we allow contracts from user swarm nodes
	isSigner := s.IsEqual(c.NodeID)
	isSelf := s.IsEqual(mod.node.Identity())
	isLocalSwarmMember := slices.ContainsFunc(mod.LocalSwarm(), s.IsEqual)

	if !(isSigner || isSelf || isLocalSwarmMember) {
		return objects.ErrPushRejected
	}

	found, err := mod.IndexSignedNodeContract(c)
	if err != nil {
		mod.log.Errorv(1, "save node contract: %v", err)
		return objects.ErrPushRejected
	}

	if !found {
		// note: temporary solution where we sync endpoints
		go func() {
			err = mod.Nodes.UpdateNodeEndpoints(mod.ctx, s, c.NodeID)
			if err != nil {
				mod.log.Error("updatingNodeEndpoint failed: %v", err)
			}
		}()

	}

	return nil
}

func (mod *Module) receiveSignedNodeContractRevocation(s *astral.Identity, r *user.SignedNodeContractRevocation, c *user.SignedNodeContract) error {
	if !slices.ContainsFunc(mod.LocalSwarm(), s.IsEqual) {
		mod.log.Errorv(1, "revoked contract from node (%v) without contract", s)
		return objects.ErrPushRejected
	}

	err := mod.ValidateNodeContractRevocation(r, c)
	if err != nil {
		mod.log.Errorv(1, "invalid node contract revocation: %v", err)
		return objects.ErrPushRejected
	}

	err = mod.SaveSignedRevocationContract(r, c)
	if err != nil {
		mod.log.Errorv(1, "save node contract revocation: %v", err)
		return objects.ErrPushRejected
	}

	return nil
}

func (mod *Module) pushActiveContract(remoteIdentity *astral.Identity) {
	contract := mod.ActiveContract()
	if contract == nil {
		return
	}

	mod.Objects.Push(mod.ctx, remoteIdentity, contract)
}

func (mod *Module) onNotification(src *astral.Identity, n *user.Notification) error {
	ac := mod.ActiveContract()
	if ac == nil {
		return objects.ErrPushRejected
	}

	if !slices.ContainsFunc(mod.ActiveNodes(ac.UserID), src.IsEqual) {
		return objects.ErrPushRejected
	}

	switch n.Event {
	case "assets":
		go mod.syncAssets(mod.ctx, src)
		return nil
	}
	return objects.ErrPushRejected
}
