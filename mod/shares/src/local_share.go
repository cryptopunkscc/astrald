package shares

import (
	"errors"
	"fmt"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/mod/sets"
	"github.com/cryptopunkscc/astrald/mod/shares"
	"time"
)

const explicitSuffix = ".explicit"

var _ shares.LocalShare = &LocalShare{}

type LocalShare struct {
	mod      *Module
	identity id.Identity
	union    sets.Union
	explicit sets.Set
	row      *dbLocalShare
}

func (mod *Module) LocalShare(identity id.Identity, create bool) (shares.LocalShare, error) {
	if share, err := mod.FindLocalShare(identity); err == nil {
		return share, nil
	}

	if create {
		return mod.CreateLocalShare(identity)
	}

	return nil, errors.New("not found")
}

func (mod *Module) CreateLocalShare(identity id.Identity) (*LocalShare, error) {
	if identity.IsZero() {
		return nil, errors.New("identity is zero")
	}

	var err error
	var share = &LocalShare{mod: mod, identity: identity}

	// create database row
	var row = dbLocalShare{
		Caller:  identity,
		SetName: share.rootSetName(),
	}

	err = mod.db.Create(&row).Error
	if err != nil {
		return nil, err
	}
	share.row = &row

	// create set structure
	share.union, err = mod.sets.CreateUnion(row.SetName)
	if err != nil {
		return nil, err
	}
	share.union.SetDisplayName(fmt.Sprintf("Data shared with {{%s}}", identity))

	share.explicit, err = mod.sets.Create(row.SetName + explicitSuffix)
	if err != nil {
		return nil, err
	}
	share.explicit.SetDisplayName(fmt.Sprintf("Data shared with {{%s}} (explicit)", identity))

	err = share.union.AddSubset(row.SetName + explicitSuffix)
	if err != nil {
		return nil, err
	}

	return share, nil
}

func (mod *Module) FindLocalShare(identity id.Identity) (*LocalShare, error) {
	var err error
	var share = &LocalShare{mod: mod, identity: identity}

	err = mod.db.
		Where("caller = ?", identity).
		First(&share.row).Error
	if err != nil {
		return nil, err
	}

	share.union, err = mod.sets.OpenUnion(share.row.SetName, false)
	if err != nil {
		return nil, err
	}

	share.explicit, err = mod.sets.Open(share.row.SetName+explicitSuffix, false)
	if err != nil {
		return nil, err
	}

	return share, nil
}

func (share *LocalShare) AddObject(dataID ...data.ID) error {
	return share.explicit.Add(dataID...)
}

func (share *LocalShare) AddSet(name ...string) error {
	return share.union.AddSubset(name...)
}

func (share *LocalShare) RemoveObject(dataID ...data.ID) error {
	return share.explicit.Remove(dataID...)
}

func (share *LocalShare) RemoveSet(name ...string) error {
	return share.union.RemoveSubset(name...)
}

func (share *LocalShare) Identity() id.Identity {
	return share.identity
}

func (share *LocalShare) Scan(opts *sets.ScanOpts) ([]*sets.Member, error) {
	return share.union.Scan(opts)
}

func (share *LocalShare) TrimmedAt() time.Time {
	info, err := share.union.Stat()
	if err != nil {
		return time.Time{}
	}
	return info.TrimmedAt
}

func (share *LocalShare) rootSetName() string {
	name := share.mod.node.Resolver().DisplayName(share.identity)

	return fmt.Sprintf("%v.%v",
		localShareSetPrefix,
		name,
	)
}
