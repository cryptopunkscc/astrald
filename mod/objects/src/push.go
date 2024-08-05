package objects

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/id"
	"github.com/cryptopunkscc/astrald/mod/objects"
)

func (mod *Module) AddReceiver(receiver objects.Receiver) error {
	if receiver == nil {
		return errors.New("receiver is nil")
	}

	return mod.receivers.Add(receiver)
}

func (mod *Module) Push(ctx context.Context, target id.Identity, obj astral.Object) (err error) {
	if target.IsEqual(mod.node.Identity()) {
		return mod.PushLocal(obj)
	}

	c, err := mod.Connect(mod.node.Identity(), target)
	if err != nil {
		return err
	}

	return c.Push(ctx, obj)
}

func (mod *Module) PushLocal(obj astral.Object) (err error) {
	objectID, err := astral.ResolveObjectID(obj)
	if err != nil {
		return err
	}

	ok := mod.pushLocal(&objects.Push{
		Source:   mod.node.Identity(),
		ObjectID: objectID,
		Object:   obj,
	})

	if !ok {
		err = errors.New("rejected")
	}

	return
}

func (mod *Module) pushLocal(push *objects.Push) (ok bool) {
	for _, r := range mod.receivers.Clone() {
		if r.ReceiveObject(push) == nil {
			ok = true
		}
	}
	if ok {
		mod.Store(push.Object)
	}
	return
}
