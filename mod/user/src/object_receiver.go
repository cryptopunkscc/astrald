package user

import (
	"context"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/nodes"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/mod/user"
)

var _ objects.Receiver = &Module{}

func (mod *Module) ReceiveObject(push *objects.SourcedObject) error {
	switch o := push.Object.(type) {
	case *user.SignedNodeContract:
		return mod.pushSignedNodeContract(push.Source, o)
	case *nodes.EventLinked:
		go mod.onNodeLinked(o)
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
	contract := mod.ActiveContract()
	if contract == nil {
		return
	}
	mod.Objects.Push(context.Background(), nil, event.NodeID, contract)
}
