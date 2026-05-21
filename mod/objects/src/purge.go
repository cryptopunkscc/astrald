package objects

import (
	"errors"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/sig"
)

func (mod *Module) purgeRepository(ctx *astral.Context, repo objects.Repository) (<-chan *astral.ObjectID, *error) {
	out := make(chan *astral.ObjectID)
	errPtr := new(error)

	go func() {
		defer close(out)

		scan, err := repo.Scan(ctx, false)
		if err != nil {
			*errPtr = err
			return
		}

		seen := map[string]bool{}
		for id := range scan {
			// note: this is n-th time i see classic dedup pattern, maybe we will create a helper
			if id == nil {
				continue
			}

			key := id.String()
			if seen[key] {
				continue
			}
			seen[key] = true

			if len(mod.Holders(id)) > 0 {
				continue
			}

			err := repo.Delete(ctx, id)
			if err != nil {
				if errors.Is(err, objects.ErrNotFound) {
					continue
				}
				*errPtr = err
				return
			}

			err = sig.Send(ctx, out, id)
			if err != nil {
				*errPtr = err
				return
			}
		}
	}()

	return out, errPtr
}
