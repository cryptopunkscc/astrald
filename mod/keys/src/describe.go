package keys

import (
	"context"
	"github.com/cryptopunkscc/astrald/lib/desc"
	"github.com/cryptopunkscc/astrald/mod/keys"
	"github.com/cryptopunkscc/astrald/object"
)

func (mod *Module) Describe(ctx context.Context, objectID object.ID, opts *desc.Opts) (descs []*desc.Desc) {
	var (
		err error
		row dbPrivateKey
	)

	err = mod.db.Where("data_id = ?", objectID).First(&row).Error
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
