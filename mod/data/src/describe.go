package data

import (
	"context"
	_data "github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/mod/data"
)

func (mod *Module) DescribeData(ctx context.Context, dataID _data.ID, opts *data.DescribeOpts) []data.Descriptor {
	if opts == nil {
		opts = &data.DescribeOpts{}
	}

	var descs []data.Descriptor

	descs = append(descs, mod.describe(dataID)...)

	for _, describer := range mod.describers.Clone() {
		var items = describer.DescribeData(ctx, dataID, nil)
		descs = append(descs, items...)
	}

	return descs
}

func (mod *Module) AddDescriber(describer data.Describer) error {
	return mod.describers.Add(describer)
}

func (mod *Module) RemoveDescriber(describer data.Describer) error {
	return mod.describers.Remove(describer)
}

func (mod *Module) describe(dataID _data.ID) []data.Descriptor {
	var descs []data.Descriptor

	row, err := mod.dbDataTypeFindByDataID(dataID.String())
	if err == nil {
		descs = append(descs, data.Descriptor{
			Type: data.TypeDescriptorType,
			Data: data.TypeDescriptor{
				Method: row.Header,
				Type:   row.Type,
			},
		})
	}

	if label := mod.GetLabel(dataID); label != "" {
		descs = append(descs, data.Descriptor{
			Type: data.LabelDescriptorType,
			Data: data.LabelDescriptor{Label: label},
		})
	}

	return descs
}
