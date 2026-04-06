package apphost

import (
	"fmt"
	"time"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/apphost"
	"github.com/cryptopunkscc/astrald/mod/objects"
)

func (mod *Module) Index(ctx *astral.Context, objectID *astral.ObjectID) (err error) {
	mod.indexMu.Lock()
	defer mod.indexMu.Unlock()

	// check if already indexed
	if _, e := mod.db.FindAppContract(objectID); e == nil {
		return nil
	}

	// load the contract from node repo
	signed, err := objects.Load[*apphost.SignedAppContract](ctx, mod.Objects.ReadDefault(), objectID)
	if err != nil {
		return fmt.Errorf("cannot load app contract: %w", err)
	}

	if !mod.isActive(signed) {
		return apphost.ErrInactiveContract
	}

	// save the contract
	err = mod.db.SaveAppContract(&dbAppContract{
		ObjectID:  objectID,
		AppID:     signed.AppID,
		HostID:    signed.HostID,
		StartsAt:  time.Time(signed.StartsAt),
		ExpiresAt: time.Time(signed.ExpiresAt),
	})
	if err != nil {
		return err
	}

	mod.Objects.Receive(&apphost.EventNewAppContract{Contract: signed}, nil)

	return
}

func (mod *Module) indexer(ctx *astral.Context) {
	ctx = ctx.ExcludeZone(astral.ZoneNetwork)

	ch, err := mod.Objects.GetRepository(objects.RepoLocal).Scan(ctx, true)
	if err != nil {
		mod.log.Error("cannot scan objects: %v", err)
		return
	}

	for objectID := range ch {
		objectType, err := mod.Objects.GetType(ctx, objectID)

		switch {
		case err != nil:
			continue
		case objectType != apphost.SignedAppContract{}.ObjectType():
			continue
		}

		_ = mod.Index(ctx, objectID)
	}

	mod.log.Logv(1, "apphost indexer finished")
}
