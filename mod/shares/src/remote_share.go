package shares

import (
	"context"
	"errors"
	"fmt"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/lib/desc"
	"github.com/cryptopunkscc/astrald/mod/sets"
	"github.com/cryptopunkscc/astrald/mod/shares"
	"strings"
	"time"
)

var _ shares.RemoteShare = &RemoteShare{}

type RemoteShare struct {
	mod    *Module
	row    *dbRemoteShare
	set    sets.Set
	caller id.Identity
	target id.Identity
}

func (mod *Module) CreateRemoteShare(caller id.Identity, target id.Identity) (*RemoteShare, error) {
	var share = &RemoteShare{
		mod:    mod,
		caller: caller,
		target: target,
	}

	var row = dbRemoteShare{
		Caller:  caller,
		Target:  target,
		SetName: share.setName(),
	}
	var err = mod.db.Create(&row).Error
	if err != nil {
		if strings.Contains(err.Error(), "constraint failed") {
			return nil, errors.New("remote share already exists")
		}
		return nil, err
	}

	share.set, err = mod.sets.CreateManaged(row.SetName, shares.RemoteSetType)
	if err != nil {
		mod.db.Delete(&row)
		return nil, fmt.Errorf("cannot create set: %w", err)
	}
	mod.sets.Network().AddSubset(row.SetName)

	return share, nil
}

func (mod *Module) FindRemoteShare(caller id.Identity, target id.Identity) (shares.RemoteShare, error) {
	return mod.findRemoteShare(caller, target)
}

func (mod *Module) findRemoteShare(caller id.Identity, target id.Identity) (*RemoteShare, error) {
	var row dbRemoteShare
	var err = mod.db.
		Where("caller = ? AND target = ?", caller, target).
		First(&row).Error
	if err != nil {
		return nil, err
	}

	var share = &RemoteShare{
		mod:    mod,
		caller: caller,
		target: target,
		row:    &row,
	}

	share.set, err = mod.sets.Open(share.row.SetName, false)
	if err != nil {
		return nil, err
	}

	return share, nil
}

func (mod *Module) FindOrCreateRemoteShare(caller id.Identity, target id.Identity) (*RemoteShare, error) {
	if share, err := mod.findRemoteShare(caller, target); err == nil {
		return share, nil
	}
	return mod.CreateRemoteShare(caller, target)
}

func (share *RemoteShare) Sync(ctx context.Context) (err error) {
	if share.target.IsEqual(share.mod.node.Identity()) {
		return errors.New("cannot sync with self")
	}

	c := NewConsumer(share.mod, share.caller, share.target)

	sync, err := c.Sync(ctx, share.LastUpdate())
	switch err {
	case nil:
	case ErrResyncRequired:
		sync, err = c.Sync(ctx, time.Time{})
		if err != nil {
			return
		}
	default:
		return
	}

	for _, update := range sync.Updates {
		if update.Removed {
			var tx = share.mod.db.Delete(&dbRemoteData{
				Caller: share.caller,
				Target: share.target,
				DataID: update.DataID,
			})

			if tx.Error != nil {
				share.mod.log.Errorv(1, "sync: error removing remote data: %v", tx.Error)
			}

			share.set.Remove(update.DataID)
		} else {
			var tx = share.mod.db.Create(&dbRemoteData{
				Caller: share.caller,
				Target: share.target,
				DataID: update.DataID,
			})
			if tx.Error != nil {
				share.mod.log.Errorv(1, "sync: error adding remote data: %v", tx.Error)
			}

			share.set.Add(update.DataID)

			// add a task to cache descriptors
			share.mod.tasks <- func(ctx context.Context) {
				_, err := share.Describe(ctx, update.DataID, &desc.Opts{
					Network: true,
				})
				if err != nil {
					share.mod.log.Errorv(2, "describe %v: %v", update.DataID, err)
				}
			}
		}
	}

	return share.SetLastUpdate(sync.Time)
}

func (share *RemoteShare) Unsync() error {
	var err error

	err = share.mod.db.
		Model(&dbRemoteData{}).
		Delete("caller = ? and target = ?", share.caller, share.target).
		Error
	if err != nil {
		return err
	}

	err = share.mod.db.Delete(&share.row).Error
	if err != nil {
		return err
	}

	return share.set.Delete()
}

func (share *RemoteShare) Reset() error {
	err := share.mod.db.
		Where("caller = ? and target = ?", share.caller, share.target).
		Delete(&dbRemoteData{}).Error
	if err != nil {
		return err
	}

	err = share.set.Clear()
	if err != nil {
		return err
	}

	return share.SetLastUpdate(time.Time{})
}

func (share *RemoteShare) Describe(ctx context.Context, dataID data.ID, opts *desc.Opts) (descs []*desc.Desc, err error) {
	cache := &DescriptorCache{mod: share.mod}

	// try cached data first
	descData, err := cache.Load(share.caller, share.target, dataID, 0)
	if err == nil {
		return addSourceToData(descData, share.target), nil
	}

	// check conditions
	if !opts.Network {
		return nil, nil
	}

	if opts.IdentityFilter != nil {
		if !opts.IdentityFilter(share.target) {
			return nil, nil
		}
	}

	// make the request
	descData, err = NewConsumer(
		share.mod,
		share.caller,
		share.target,
	).Describe(ctx, dataID, opts)
	if err != nil {
		return
	}

	// cache results
	err = cache.Store(share.caller, share.target, dataID, descData)
	if err != nil {
		share.mod.log.Error("error storing cache: %v", err)
	}

	return addSourceToData(descData, share.target), nil
}

func addSourceToData(list []desc.Data, source id.Identity) (descs []*desc.Desc) {
	for _, item := range list {
		descs = append(descs, &desc.Desc{
			Source: source,
			Data:   item,
		})
	}
	return
}

func (share *RemoteShare) Scan(opts *sets.ScanOpts) ([]*sets.Member, error) {
	return share.set.Scan(opts)
}

func (share *RemoteShare) LastUpdate() time.Time {
	return share.row.LastUpdate
}

func (share *RemoteShare) SetLastUpdate(t time.Time) error {
	share.row.LastUpdate = t
	return share.mod.db.Save(&share.row).Error
}

func (share *RemoteShare) setName() string {
	return fmt.Sprintf("%v.%v.%v",
		remoteShareSetPrefix,
		share.mod.node.Resolver().DisplayName(share.caller),
		share.mod.node.Resolver().DisplayName(share.target),
	)
}
