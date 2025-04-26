package objects

import (
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

func (mod *Module) Receive(obj astral.Object, source *astral.Identity) (err error) {
	if source.IsZero() {
		source = mod.node.Identity()
	}

	ok := mod.receive(&objects.SourcedObject{
		Source: mod.node.Identity(),
		Object: obj,
	})

	if !ok {
		err = errors.New("rejected")
	}

	return
}

func (mod *Module) receive(push *objects.SourcedObject) (ok bool) {
	for _, r := range mod.receivers.Clone() {
		if r.ReceiveObject(push) == nil {
			ok = true
		}
	}
	if ok {
		mod.Save(push.Object)
	}
	return
}
