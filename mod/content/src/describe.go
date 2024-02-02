package content

import (
	"context"
	"github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/mod/content"
)

func (mod *Module) Describe(ctx context.Context, dataID data.ID, opts *content.DescribeOpts) []*content.Descriptor {
	if opts == nil {
		opts = &content.DescribeOpts{}
	}

	var descs []*content.Descriptor

	descs = append(descs, mod.describe(dataID)...)

	for _, describer := range mod.describers.Clone() {
		var items = describer.Describe(ctx, dataID, opts)
		descs = append(descs, items...)
	}

	return descs
}

func (mod *Module) AddDescriber(describer content.Describer) error {
	return mod.describers.Add(describer)
}

func (mod *Module) RemoveDescriber(describer content.Describer) error {
	return mod.describers.Remove(describer)
}

func (mod *Module) describe(dataID data.ID) []*content.Descriptor {
	var descs []*content.Descriptor
	var err error
	var row dbDataType

	err = mod.db.Where("data_id = ?", dataID).First(&row).Error
	if err == nil {
		descs = append(descs, &content.Descriptor{
			Source: mod.node.Identity(),
			Info: content.TypeDescriptor{
				Method: row.Method,
				Type:   row.Type,
			},
		})
	}

	if label := mod.GetLabel(dataID); label != "" {
		descs = append(descs, &content.Descriptor{
			Source: mod.node.Identity(),
			Info:   content.LabelDescriptor{Label: label},
		})
	}

	return descs
}
