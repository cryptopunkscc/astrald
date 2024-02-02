package keys

import (
	"context"
	"github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/mod/content"
	"github.com/cryptopunkscc/astrald/mod/keys"
)

func (mod *Module) Describe(ctx context.Context, dataID data.ID, opts *content.DescribeOpts) (desc []*content.Descriptor) {
	var (
		err error
		row dbPrivateKey
	)

	err = mod.db.Where("data_id = ?", dataID).First(&row).Error
	if err != nil {
		return
	}

	desc = append(desc, &content.Descriptor{
		Source: mod.node.Identity(),
		Info: keys.KeyDescriptor{
			KeyType:   row.Type,
			PublicKey: row.PublicKey,
		},
	})

	return
}
