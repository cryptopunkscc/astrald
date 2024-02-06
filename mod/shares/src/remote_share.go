package shares

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/mod/content"
	"github.com/cryptopunkscc/astrald/mod/sets"
	"github.com/cryptopunkscc/astrald/mod/shares"
	"github.com/cryptopunkscc/astrald/net"
	"io"
	"strconv"
	"strings"
	"time"
)

var _ sets.Set = &RemoteShare{}
var _ shares.RemoteShare = &RemoteShare{}

type RemoteShare struct {
	mod    *Module
	caller id.Identity
	target id.Identity
	set    sets.Editor
	row    *dbRemoteShare
}

func (mod *Module) setOpener(name string) (sets.Set, error) {
	after, ok := strings.CutPrefix(name, remoteShareSetPrefix+".")
	if !ok {
		return nil, errors.New("invalid set name")
	}

	idstr := strings.Split(after, ".")
	if len(idstr) != 2 {
		return nil, errors.New("invalid set name")
	}

	caller, err := id.ParsePublicKeyHex(idstr[0])
	if err != nil {
		return nil, err
	}

	target, err := id.ParsePublicKeyHex(idstr[1])
	if err != nil {
		return nil, err
	}

	var share = &RemoteShare{
		mod:    mod,
		caller: caller,
		target: target,
	}

	share.set, err = mod.sets.Edit(name)
	if err != nil {
		return nil, err
	}

	return share, nil
}

func (mod *Module) CreateRemoteShare(caller id.Identity, target id.Identity) (*RemoteShare, error) {
	var share = &RemoteShare{mod: mod, caller: caller, target: target}

	var row = dbRemoteShare{
		Caller: caller,
		Target: target,
	}

	var err = mod.db.Create(&row).Error
	if err != nil {
		return nil, err
	}
	share.row = &row

	setName := share.setName()

	_, err = mod.sets.Create(setName, shares.SetType)
	if err != nil {
		return nil, err
	}

	share.set, err = mod.sets.Edit(setName)
	if err != nil {
		return nil, err
	}

	mod.remoteShares.Add(setName)
	mod.sets.SetVisible(setName, true)
	mod.sets.SetDescription(setName,
		fmt.Sprintf(
			"Data shared with %s by %s",
			mod.node.Resolver().DisplayName(caller),
			mod.node.Resolver().DisplayName(target),
		),
	)

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

	var share = &RemoteShare{mod: mod, caller: caller, target: target, row: &row}
	share.set, err = mod.sets.Edit(share.setName())
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

func (share *RemoteShare) Sync() error {
	if share.target.IsEqual(share.mod.node.Identity()) {
		return errors.New("cannot sync with self")
	}

	remoteShare, err := share.mod.FindOrCreateRemoteShare(share.caller, share.target)
	if err != nil {
		return err
	}

	var timestamp = "0"
	if !remoteShare.row.LastUpdate.IsZero() {
		timestamp = strconv.FormatInt(remoteShare.row.LastUpdate.UnixNano(), 10)
	}

	var query = net.NewQuery(share.caller, share.target, syncServicePrefix+timestamp)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	conn, err := net.Route(ctx, share.mod.node.Router(), query)
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
		case opDone: // done
			var timestamp int64
			err = cslq.Decode(conn, "q", &timestamp)
			if err != nil {
				return err
			}

			return remoteShare.SetLastUpdate(time.Unix(0, timestamp))

		case opAdd: // add
			var dataID data.ID
			err = cslq.Decode(conn, "v", &dataID)
			if err != nil {
				return err
			}

			var tx = share.mod.db.Create(&dbRemoteData{
				Caller: share.caller,
				Target: share.target,
				DataID: dataID,
			})
			if tx.Error != nil {
				share.mod.log.Errorv(1, "sync: error adding remote data: %v", tx.Error)
			}

			remoteShare.set.Add(dataID)

			// cache descriptors
			share.mod.tasks <- func(ctx context.Context) {
				_, err := remoteShare.Describe(ctx, dataID, &content.DescribeOpts{
					Network: true,
				})
				if err != nil {
					share.mod.log.Errorv(2, "describe %v: %v", dataID, err)
				}
			}

		case opRemove: // remove
			var dataID data.ID
			err = cslq.Decode(conn, "v", &dataID)
			if err != nil {
				return err
			}

			var tx = share.mod.db.Delete(&dbRemoteData{
				Caller: share.caller,
				Target: share.target,
				DataID: dataID,
			})

			if tx.Error != nil {
				share.mod.log.Errorv(1, "sync: error removing remote data: %v", tx.Error)
			}

			remoteShare.set.Remove(dataID)

		case opResync:
			conn.Close()
			err = share.Reset()
			if err != nil {
				return err
			}
			return share.Sync()

		case opNotFound:
			return errors.New("remote share not found")

		default:
			return errors.New("protocol error")
		}
	}

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

	err = share.set.Reset()
	if err != nil {
		return err
	}

	return share.SetLastUpdate(time.Time{})
}

func (share *RemoteShare) Describe(ctx context.Context, dataID data.ID, opts *content.DescribeOpts) ([]*content.Descriptor, error) {
	var list []*content.Descriptor
	var rawJSON []byte

	var row dbRemoteDesc
	err := share.mod.db.
		Where("caller = ? AND target = ? AND data_id = ?", share.caller, share.target, dataID).
		First(&row).Error

	if err == nil {
		rawJSON = []byte(row.Desc)
	} else {
		if !opts.Network {
			return nil, nil
		}
		if opts.IdentityFilter != nil {
			if !opts.IdentityFilter(share.target) {
				return nil, nil
			}
		}

		var query = net.NewQuery(
			share.caller,
			share.target,
			describeServiceName+"?id="+dataID.String(),
		)

		conn, err := net.Route(ctx, share.mod.node.Router(), query)
		if err != nil {
			return nil, err
		}
		defer conn.Close()

		rawJSON, err = io.ReadAll(conn)
		if err != nil {
			return nil, err
		}

		share.mod.db.Create(&dbRemoteDesc{
			Caller: share.caller,
			Target: share.target,
			DataID: dataID,
			Desc:   string(rawJSON),
		})
	}

	var j []JSONDescriptor
	err = json.Unmarshal(rawJSON, &j)
	if err != nil {
		return nil, err
	}

	for _, d := range j {
		var proto = share.mod.content.UnmarshalDescriptor(d.Type, d.Info)
		if proto == nil {
			continue
		}

		list = append(list, &content.Descriptor{
			Source: share.target,
			Data:   proto,
		})
	}

	return list, nil
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

func (share *RemoteShare) Info() (*sets.Stat, error) {
	return &sets.Stat{
		Name:        share.setName(),
		Type:        "share",
		Size:        0,
		Visible:     false,
		Description: "",
		CreatedAt:   time.Time{},
		TrimmedAt:   share.set.TrimmedAt(),
	}, nil
}

func (share *RemoteShare) setName() string {
	return fmt.Sprintf("%v.%v.%v",
		remoteShareSetPrefix,
		share.caller.PublicKeyHex(),
		share.target.PublicKeyHex(),
	)
}
