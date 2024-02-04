package shares

import (
	"context"
	"github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/mod/content"
)

var _ content.Describer = &Module{}

func (mod *Module) Describe(ctx context.Context, dataID data.ID, opts *content.DescribeOpts) []*content.Descriptor {
	var list []*content.Descriptor
	var err error
	var rows []*dbRemoteData

	err = mod.db.Where("data_id = ?", dataID).Find(&rows).Error
	if err != nil {
		return nil
	}

	for _, row := range rows {
		share, err := mod.findRemoteShare(row.Caller, row.Target)
		if err != nil {
			continue
		}

		res, err := share.Describe(ctx, dataID, opts)
		if err != nil {
			continue
		}

		list = append(list, res...)
	}

	return list
}
