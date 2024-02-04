package shares

import (
	"fmt"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/mod/sets"
	"time"
)

type LocalShare struct {
	identity id.Identity
	union    sets.Union
	basic    sets.Basic
}

func (mod *Module) FindOrCreateLocalShare(identity id.Identity) (*LocalShare, error) {
	if share, err := mod.FindLocalShare(identity); err == nil {
		return share, nil
	}

	return mod.CreateLocalShare(identity)
}

func (mod *Module) CreateLocalShare(identity id.Identity) (*LocalShare, error) {
	var err error
	var share = &LocalShare{identity: identity}

	share.union, err = mod.sets.CreateUnion(share.unionSetName())
	if err != nil {
		return nil, err
	}

	share.basic, err = mod.sets.CreateBasic(share.basicSetName())
	if err != nil {
		return nil, err
	}

	err = share.union.Add(share.basicSetName())
	if err != nil {
		return nil, err
	}

	mod.sets.SetVisible(share.unionSetName(), true)
	mod.sets.SetDescription(share.unionSetName(),
		fmt.Sprintf(
			"Data shared with %s",
			mod.node.Resolver().DisplayName(identity),
		),
	)

	return share, nil
}

func (mod *Module) FindLocalShare(identity id.Identity) (*LocalShare, error) {
	var err error
	var share = &LocalShare{identity: identity}

	share.union, err = sets.Open[sets.Union](mod.sets, share.unionSetName())
	if err != nil {
		return nil, err
	}

	share.basic, err = sets.Open[sets.Basic](mod.sets, share.basicSetName())
	if err != nil {
		return nil, err
	}

	return share, nil
}

func (share *LocalShare) AddData(dataID ...data.ID) error {
	return share.basic.Add(dataID...)
}

func (share *LocalShare) AddSet(name ...string) error {
	return share.union.Add(name...)
}

func (share *LocalShare) RemoveData(dataID ...data.ID) error {
	return share.basic.Remove(dataID...)
}

func (share *LocalShare) RemoveSet(name ...string) error {
	return share.union.Remove(name...)
}

func (share *LocalShare) Scan(opts *sets.ScanOpts) ([]*sets.Member, error) {
	return share.union.Scan(opts)
}

func (share *LocalShare) TrimmedAt() time.Time {
	info, err := share.union.Info()
	if err != nil {
		return time.Time{}
	}
	return info.TrimmedAt
}

func (share *LocalShare) basicSetName() string {
	return fmt.Sprintf("%v.%v.set",
		localShareSetPrefix,
		share.identity.PublicKeyHex(),
	)
}

func (share *LocalShare) unionSetName() string {
	return fmt.Sprintf("%v.%v",
		localShareSetPrefix,
		share.identity.PublicKeyHex(),
	)
}
