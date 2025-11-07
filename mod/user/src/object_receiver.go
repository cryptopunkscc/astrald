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
	case *apphost.AppContract:
		if !slices.ContainsFunc(mod.LocalSwarm(), o.HostID.IsEqual) {
			break
		}

		drop.Accept(true)

	case *nodes.EventLinked:
		go mod.onNodeLinked(o)

		drop.Accept(false)
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
		case *nodes.StreamClosedEvent:
			if slices.ContainsFunc(mod.LocalSwarm(), e.With.IsEqual) {
				mod.Scheduler.Schedule(mod.ctx, mod.NewEnsureConnectivityAction())
				drop.Accept(false)
			}
		}
	}

	return nil
}

func (mod *Module) receiveSignedNodeContract(s *astral.Identity, c *user.SignedNodeContract) error {
	// reject contracts coming from neither the signing node nor local node
	if !(s.IsEqual(c.NodeID) || s.IsEqual(mod.node.Identity())) {
		return objects.ErrPushRejected
	}

	err := mod.SaveSignedNodeContract(c)
	if err != nil {
		mod.log.Errorv(1, "save node contract: %v", err)
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
		go mod.SyncAssets(mod.ctx, src)
		return nil
	}
	return objects.ErrPushRejected
}
