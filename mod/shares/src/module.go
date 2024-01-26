package shares

import (
	"context"
	"errors"
	"fmt"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/index"
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
	index       index.Module
	notify      sig.Set[string]
}

const localShareIndexPrefix = "mod.shares.local"
const remoteShareIndexPrefix = "mod.shares.remote"
const publicIndexName = "mod.shares.public"
const setSuffix = ".set"
const resyncInterval = time.Hour
const resyncAge = 5 * time.Minute

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

	tasks.Group(
		NewReadService(mod),
		NewSyncService(mod),
		NewNotifyService(mod),
	).Run(ctx)

	<-ctx.Done()

	return nil
}

func (mod *Module) Sync(caller id.Identity, target id.Identity) error {
	if target.IsEqual(mod.node.Identity()) {
		return errors.New("cannot sync with self")
	}

	var row dbRemoteShare
	var timestamp = "0"
	var remoteIndexName = mod.remoteShareIndexName(caller, target)
	var fetchTx = mod.db.Where("caller = ? and target = ?", caller, target).First(&row)
	if fetchTx.Error == nil {
		timestamp = strconv.FormatInt(row.LastUpdate.UnixNano(), 10)
	} else {
		mod.index.CreateIndex(remoteIndexName, index.TypeSet)
	}

	var query = net.NewQuery(caller, target, syncServicePrefix+timestamp)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	conn, err := net.Route(ctx, mod.node.Router(), query)
	if err != nil {
		return err
	}
	defer conn.Close()

	for {
		var op byte
		err = cslq.Decode(conn, "c", &op)
		if err != nil {
			return err
		}

		switch op {
		case 0: // done
			var timestamp int64
			err = cslq.Decode(conn, "q", &timestamp)

			row.Caller = caller
			row.Target = target
			row.LastUpdate = time.Unix(0, timestamp)

			var tx = mod.db.Save(&row)
			if tx.Error != nil {
				mod.log.Errorv(1, "sync: error updating share: %v", tx.Error)
			}

			return tx.Error

		case 1: // add
			var dataID data.ID
			err = cslq.Decode(conn, "v", &dataID)
			if err != nil {
				return err
			}

			var tx = mod.db.Create(&dbRemoteData{
				Caller: caller,
				Target: target,
				DataID: dataID,
			})
			if tx.Error != nil {
				mod.log.Errorv(1, "sync: error adding remote data: %v", tx.Error)
			}

			mod.index.AddToSet(remoteIndexName, dataID)

		case 2: // remove
			var dataID data.ID
			err = cslq.Decode(conn, "v", &dataID)
			if err != nil {
				return err
			}

			var tx = mod.db.Delete(&dbRemoteData{
				Caller: caller,
				Target: target,
				DataID: dataID,
			})

			if tx.Error != nil {
				mod.log.Errorv(1, "sync: error removing remote data: %v", tx.Error)
			}

			mod.index.RemoveFromSet(remoteIndexName, dataID)

		default:
			return errors.New("protocol error")
		}
	}

}

func (mod *Module) Unsync(caller id.Identity, target id.Identity) error {
	var share dbRemoteShare
	var err = mod.db.Where("caller = ? and target = ?", caller, target).First(&share).Error
	if err != nil {
		return err
	}

	err = mod.db.
		Model(&dbRemoteData{}).
		Delete("caller = ? and target = ?", caller, target).
		Error

	if err != nil {
		return err
	}

	err = mod.db.Delete(&share).Error
	if err == nil {
		mod.index.DeleteIndex(mod.remoteShareIndexName(caller, target))
	}
	return err
}

func (mod *Module) Read(dataID data.ID, opts *storage.ReadOpts) (storage.DataReader, error) {
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

		return &RemoteDataReader{ReadCloser: conn}, nil
	}

	return nil, storage.ErrNotFound
}

func (mod *Module) RemoteShares() ([]shares.RemoteShare, error) {
	var rows []dbRemoteShare

	var err = mod.db.Find(&rows).Error
	if err != nil {
		return nil, err
	}

	var list []shares.RemoteShare
	for _, row := range rows {
		list = append(list, shares.RemoteShare{
			Caller: row.Caller,
			Target: row.Target,
		})
	}

	return list, nil
}

func (mod *Module) ListRemote(caller id.Identity, target id.Identity) ([]data.ID, error) {
	var rows []dbRemoteData
	var err = mod.db.
		Where("caller = ? and target = ?", caller, target).
		Find(&rows).
		Error
	if err != nil {
		return nil, err
	}

	return sig.MapSlice(rows, func(row dbRemoteData) (data.ID, error) {
		return row.DataID, nil
	})
}

func (mod *Module) LastSynced(caller id.Identity, target id.Identity) (time.Time, error) {
	var row dbRemoteShare
	var err = mod.db.
		Where("caller = ? and target = ?", caller, target).
		First(&row).
		Error
	if err != nil {
		return time.Time{}, err
	}

	return row.LastUpdate, nil
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
		err := mod.Sync(row.Caller, row.Target)
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

func (mod *Module) makeLocalIndexFor(identity id.Identity) error {
	unionName := mod.localShareIndexName(identity)

	info, err := mod.index.IndexInfo(unionName)
	if err == nil {
		if info.Type != index.TypeUnion {
			return fmt.Errorf("index %s is not a union", unionName)
		}
		return nil
	}

	var setName = unionName + setSuffix

	_, err = mod.index.CreateIndex(unionName, index.TypeUnion)
	if err != nil {
		return err
	}

	_, err = mod.index.CreateIndex(setName, index.TypeSet)
	if err != nil {
		return err
	}

	err = mod.index.AddToUnion(unionName, publicIndexName)
	if err != nil {
		return err
	}

	return mod.index.AddToUnion(unionName, setName)
}

func (mod *Module) localShareIndexName(identity id.Identity) string {
	return fmt.Sprintf("%v.%v",
		localShareIndexPrefix,
		identity.PublicKeyHex(),
	)
}

func (mod *Module) remoteShareIndexName(guest id.Identity, host id.Identity) string {
	return fmt.Sprintf("%v.%v.%v",
		remoteShareIndexPrefix,
		guest.PublicKeyHex(),
		host.PublicKeyHex(),
	)
}

func (mod *Module) addToLocalShareIndex(identity id.Identity, dataID data.ID) error {
	var err = mod.makeLocalIndexFor(identity)
	if err != nil {
		return err
	}

	return mod.index.AddToSet(mod.localShareIndexName(identity)+setSuffix, dataID)
}

func (mod *Module) removeFromLocalShareIndex(identity id.Identity, dataID data.ID) error {
	return mod.index.RemoveFromSet(mod.localShareIndexName(identity)+setSuffix, dataID)
}

func (mod *Module) localShareIndexContains(identity id.Identity, dataID data.ID) (bool, error) {
	contains, err := mod.index.Contains(mod.localShareIndexName(identity), dataID)
	if errors.Is(err, index.ErrIndexNotFound) {
		return false, nil
	}
	return contains, err
}
