package user

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/nodes"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/mod/user"
	"slices"
)

var _ objects.Receiver = &Module{}

func (mod *Module) ReceiveObject(push *objects.SourcedObject) error {
	switch o := push.Object.(type) {
	case *user.SignedNodeContract:
		return mod.pushSignedNodeContract(push.Source, o)
	case *nodes.EventLinked:
		go mod.onNodeLinked(o)
	case *user.Notification:
		return mod.onNotification(push.Source, o)
	}

	return objects.ErrPushRejected
}

func (mod *Module) pushSignedNodeContract(s *astral.Identity, c *user.SignedNodeContract) error {
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

func (mod *Module) onNodeLinked(event *nodes.EventLinked) {
	ctx := astral.NewContext(nil).WithIdentity(mod.node.Identity())
	contract := mod.ActiveContract()
	if contract == nil {
		return
	}
	mod.Objects.Push(ctx, event.NodeID, contract)
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
