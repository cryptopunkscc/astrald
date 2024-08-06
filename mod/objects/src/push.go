package objects

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/objects"
)

func (mod *Module) AddReceiver(receiver objects.Receiver) error {
	if receiver == nil {
		return errors.New("receiver is nil")
	}

	return mod.receivers.Add(receiver)
}

func (mod *Module) Push(ctx context.Context, source *astral.Identity, target *astral.Identity, obj astral.Object) (err error) {
	if source.IsZero() {
		source = mod.node.Identity()
	}

	if target.IsEqual(mod.node.Identity()) {
		return mod.PushLocal(source, obj)
	}

	c, err := mod.Connect(source, target)
	if err != nil {
		return err
	}

	return c.Push(ctx, obj)
}

func (mod *Module) PushLocal(source *astral.Identity, obj astral.Object) (err error) {
	if source.IsZero() {
		source = mod.node.Identity()
	}

	ok := mod.pushLocal(&objects.SourcedObject{
		Source: mod.node.Identity(),
		Object: obj,
	})

	if !ok {
		err = errors.New("rejected")
	}

	return
}

func (mod *Module) pushLocal(push *objects.SourcedObject) (ok bool) {
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
