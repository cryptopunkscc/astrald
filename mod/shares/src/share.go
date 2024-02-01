package shares

import (
	"fmt"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/mod/sets"
	"time"
)

type Share struct {
	identity id.Identity
	union    sets.Union
	basic    sets.Basic
}

func (mod *Module) FindOrCreateShare(identity id.Identity) (*Share, error) {
	if share, err := mod.FindShare(identity); err == nil {
		return share, nil
	}

	return mod.CreateShare(identity)
}

func (mod *Module) CreateShare(identity id.Identity) (*Share, error) {
	var err error
	var share = &Share{identity: identity}

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

func (mod *Module) FindShare(identity id.Identity) (*Share, error) {
	var err error
	var share = &Share{identity: identity}

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

func (share *Share) AddData(dataID ...data.ID) error {
	return share.basic.Add(dataID...)
}

func (share *Share) AddSet(name ...string) error {
	return share.union.Add(name...)
}

func (share *Share) RemoveData(dataID ...data.ID) error {
	return share.basic.Remove(dataID...)
}

func (share *Share) RemoveSet(name ...string) error {
	return share.union.Remove(name...)
}

func (share *Share) Scan(opts *sets.ScanOpts) ([]*sets.Member, error) {
	return share.union.Scan(opts)
}

func (share *Share) TrimmedAt() time.Time {
	info, err := share.union.Info()
	if err != nil {
		return time.Time{}
	}
	return info.TrimmedAt
}

func (share *Share) basicSetName() string {
	return fmt.Sprintf("%v.%v.set",
		localShareSetPrefix,
		share.identity.PublicKeyHex(),
	)
}

func (share *Share) unionSetName() string {
	return fmt.Sprintf("%v.%v",
		localShareSetPrefix,
		share.identity.PublicKeyHex(),
	)
}
