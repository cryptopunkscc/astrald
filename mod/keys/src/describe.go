package keys

import (
	"context"
	"github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/lib/desc"
	"github.com/cryptopunkscc/astrald/mod/keys"
)

func (mod *Module) Describe(ctx context.Context, dataID data.ID, opts *desc.Opts) (descs []*desc.Desc) {
	var (
		err error
		row dbPrivateKey
	)

	err = mod.db.Where("data_id = ?", dataID).First(&row).Error
	if err != nil {
		return
	}

	descs = append(descs, &desc.Desc{
		Source: mod.node.Identity(),
		Data: keys.KeyDesc{
			KeyType:   row.Type,
			PublicKey: row.PublicKey,
		},
	})

	return
}
