package infra

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/infra"
)

var _ infra.Announcer = &Infra{}

func (i *Infra) Announce(ctx context.Context, id id.Identity) error {
	var count int

	for network := range i.Networks() {
		announcer, ok := network.(infra.Announcer)
		if !ok {
			continue
		}

		if announcer.Announce(ctx, id) == nil {
			count++
		}
	}

	if count == 0 {
		return errors.New("failed to announce on any network")
	}

	return nil
}
