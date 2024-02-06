package shares

import (
	"context"
	"github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/content"
	"github.com/cryptopunkscc/astrald/mod/sets"
	"github.com/cryptopunkscc/astrald/mod/shares"
	"github.com/cryptopunkscc/astrald/mod/storage"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/node/assets"
	"github.com/cryptopunkscc/astrald/node/router"
	"github.com/cryptopunkscc/astrald/sig"
	"github.com/cryptopunkscc/astrald/tasks"
	"gorm.io/gorm"
	"strconv"
	"time"
)

var _ shares.Module = &Module{}

type Module struct {
	config      Config
	node        node.Node
	log         *log.Logger
	assets      assets.Assets
	db          *gorm.DB
	authorizers sig.Set[shares.Authorizer]
	storage     storage.Module
	sets        sets.Module
	content     content.Module
	notify      sig.Map[string, *Notification]
	tasks       chan func(ctx context.Context)
}

const localShareSetPrefix = ".shares.local"
const remoteShareSetPrefix = ".shares.remote"
const resyncInterval = time.Hour
const resyncAge = 5 * time.Minute
const workers = 8

func (mod *Module) Run(ctx context.Context) error {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(resyncInterval):
				mod.resync(resyncAge)
			}
		}
	}()

	for i := 0; i < workers; i++ {
		go func() {
			for {
				select {
				case <-ctx.Done():
					return
				case task := <-mod.tasks:
					task(ctx)
				}
			}
		}()
	}

	tasks.Group(
		NewReadService(mod),
		NewSyncService(mod),
		NewNotifyService(mod),
		NewDescribeService(mod),
	).Run(ctx)

	<-ctx.Done()

	return nil
}

func (mod *Module) Open(dataID data.ID, opts *storage.OpenOpts) (storage.Reader, error) {
	if !opts.Network {
		return nil, storage.ErrNotFound
	}

	var rows []dbRemoteData

	var tx = mod.db.Where("data_id = ?", dataID).Find(&rows)
	if tx.Error != nil {
		return nil, storage.ErrNotFound
	}

	params := map[string]string{
		"id": dataID.String(),
	}

	if opts.Offset != 0 {
		params["offset"] = strconv.FormatUint(opts.Offset, 10)
	}

	var query = router.FormatQuery(readServiceName, params)
	for _, row := range rows {
		// apply identity filter if provided
		if opts.IdentityFilter != nil {
			if !opts.IdentityFilter(row.Target) {
				continue
			}
		}

		var q = net.NewQuery(
			row.Caller,
			row.Target,
			query,
		)

		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		conn, err := net.Route(ctx, mod.node.Router(), q)
		if err != nil {
			continue
		}

		return &RemoteDataReader{
			caller:     row.Caller,
			target:     row.Target,
			mod:        mod,
			dataID:     dataID,
			ReadCloser: conn,
		}, nil
	}

	return nil, storage.ErrNotFound
}

func (mod *Module) resync(minAge time.Duration) error {
	var rows []dbRemoteShare
	var err = mod.db.
		Where("last_update < ?", time.Now().Add(-minAge)).
		Find(&rows).
		Error
	if err != nil {
		return err
	}

	for _, row := range rows {
		share, err := mod.FindRemoteShare(row.Caller, row.Target)
		if err != nil {
			mod.log.Errorv(1, "find %v@%v error: %v",
				row.Caller,
				row.Target,
				err,
			)
		}

		err = share.Sync()
		if err != nil {
			mod.log.Errorv(1, "sync %v@%v error: %v",
				row.Caller,
				row.Target,
				err,
			)
		}
	}

	return nil
}
