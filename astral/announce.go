package astral

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/auth/id"
)

func (astral *Astral) Announce(ctx context.Context, id id.Identity) error {
	var count int

	for network := range astral.Networks() {
		if network.Announce(ctx, id) == nil {
			count++
		}
	}

	if count == 0 {
		return errors.New("failed to announce on any network")
	}

	return nil
}
