package objects

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/sig"
	"sync"
)

type Drop struct {
	mod      *Module
	senderID *astral.Identity
	object   astral.Object
	repo     objects.Repository
	accepted sig.Value[bool]
	saved    bool
	mu       sync.Mutex
}

func (drop *Drop) SenderID() *astral.Identity {
	return drop.senderID
}

func (drop *Drop) Object() astral.Object {
	return drop.object
}

func (drop *Drop) Accept(save bool) error {
	drop.accepted.Set(true)
	if !save {
		return nil
	}

	drop.mu.Lock()
	defer drop.mu.Unlock()

	if drop.saved {
		return nil
	}

	ctx := astral.NewContext(nil)

	_, err := objects.Save(ctx, drop.object, drop.repo)
	if err != nil {
		drop.mod.log.Error("error saving received object: %v", err)
	} else {
		drop.saved = true
	}

	return err
}
