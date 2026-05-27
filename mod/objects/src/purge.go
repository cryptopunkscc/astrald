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

		// flush pending reads so the read order reflects the latest accesses
		err := mod.objectsReadsJournal.Flush()
		if err != nil {
			mod.log.Error("object reads journal: purge flush: %v", err)
		}

		// purge in read order: oldest-read objects first, paged via keyset cursor
		var after *readCursor
		for {
			select {
			case <-ctx.Done():
				*errPtr = ctx.Err()
				return
			default:
			}

			ids, next, err := mod.db.ListReadOldest(after, 256)
			if err != nil {
				*errPtr = err
				return
			}

			for _, id := range ids {
				if len(mod.Holders(id)) > 0 {
					continue
				}

				err := repo.Delete(ctx, id)
				switch {
				case err == nil:
					err = sig.Send(ctx, out, id)
					if err != nil {
						*errPtr = err
						return
					}

				case errors.Is(err, objects.ErrNotFound):
					derr := mod.db.DeleteObjectCacheByID(id)
					if derr != nil {
						mod.log.Error("purge: drop stale cache row %v: %v", id, derr)
					}
					continue
				case errors.Is(err, errors.ErrUnsupported):
					continue
				default:
					*errPtr = err
					return
				}
			}

			if next == nil {
				break
			}
			after = next
		}
	}()

	return out, errPtr
}
