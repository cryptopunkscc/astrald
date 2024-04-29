package shares

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/lib/desc"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/mod/sets"
	"github.com/cryptopunkscc/astrald/mod/shares"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/node/assets"
	"github.com/cryptopunkscc/astrald/object"
	"github.com/cryptopunkscc/astrald/sig"
	"gorm.io/gorm"
	"sync"
	"time"
)

const localShareSetPrefix = ".shares.local"
const remoteShareSetPrefix = ".shares.remote"
const resyncInterval = time.Hour
const resyncAge = 5 * time.Minute
const workers = 8

var _ shares.Module = &Module{}
var _ objects.Describer = &Module{}

type Module struct {
	config      Config
	node        node.Node
	log         *log.Logger
	assets      assets.Assets
	db          *gorm.DB
	authorizers sig.Set[DataAuthorizer]
	objects     objects.Module
	sets        sets.Module
	notify      sig.Map[string, *Notification]
	tasks       chan func(ctx context.Context)

	shares *Provider
}

type Notification struct {
	*Module
	Identity id.Identity
	NotifyAt time.Time
	sync.Mutex
}

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

	mod.shares = NewProvider(mod)

	err := mod.node.LocalRouter().AddRoute("shares.*", mod.shares)
	if err != nil {
		return err
	}

	<-ctx.Done()

	return nil
}

func (mod *Module) Open(objectID object.ID, opts *objects.OpenOpts) (objects.Reader, error) {
	if !opts.Network {
		return nil, objects.ErrNotFound
	}

	var rows []dbRemoteData

	var tx = mod.db.Where("data_id = ?", objectID).Find(&rows)
	if tx.Error != nil {
		return nil, objects.ErrNotFound
	}

	for _, row := range rows {
		remoteObjects, err := mod.objects.Connect(row.Caller, row.Target)
		if err != nil {
			continue
		}

		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		conn, err := remoteObjects.Open(ctx, objectID, opts)
		if err != nil {
			continue
		}

		return &RemoteDataReader{
			caller:     row.Caller,
			target:     row.Target,
			mod:        mod,
			objectID:   objectID,
			ReadCloser: conn,
		}, nil
	}

	return nil, objects.ErrNotFound
}

func (mod *Module) Describe(ctx context.Context, objectID object.ID, opts *desc.Opts) []*desc.Desc {
	var list []*desc.Desc
	var err error
	var rows []*dbRemoteData

	err = mod.db.Where("data_id = ?", objectID).Find(&rows).Error
	if err != nil {
		return nil
	}

	for _, row := range rows {
		share, err := mod.findRemoteShare(row.Caller, row.Target)
		if err != nil {
			continue
		}

		res, err := share.Describe(ctx, objectID, opts)
		if err != nil {
			continue
		}

		list = append(list, res...)
	}

	return list
}

func (mod *Module) Authorize(identity id.Identity, objectID object.ID) error {
	for _, authorizer := range mod.authorizers.Clone() {
		var err = authorizer.Authorize(identity, objectID)
		switch {
		case err == nil:
			return nil
		case errors.Is(err, shares.ErrDenied):
		default:
			return err
		}
	}
	return shares.ErrDenied
}

func (mod *Module) openExportSet(identity id.Identity) (sets.Set, error) {
	return mod.sets.Open("export:"+identity.PublicKeyHex(), false)
}

func (mod *Module) addAuthorizer(authorizer DataAuthorizer) error {
	return mod.authorizers.Add(authorizer)
}

func (mod *Module) removeAuthorizer(authorizer DataAuthorizer) error {
	return mod.authorizers.Remove(authorizer)
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

		err = share.Sync(context.TODO())
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
