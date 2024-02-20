package content

import (
	"context"
	"encoding/json"
	"github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/lib/desc"
	"github.com/cryptopunkscc/astrald/mod/content"
	"reflect"
)

func (mod *Module) Describe(ctx context.Context, dataID data.ID, opts *desc.Opts) []*desc.Desc {
	var list = []desc.Describer[data.ID]{
		desc.Func(mod.describe),
	}
	for _, d := range mod.describers.Clone() {
		list = append(list, d)
	}

	return desc.Collect(ctx, dataID, opts, list...)
}

func (mod *Module) AddDescriber(describer content.Describer) error {
	return mod.describers.Add(describer)
}

func (mod *Module) RemoveDescriber(describer content.Describer) error {
	return mod.describers.Remove(describer)
}

func (mod *Module) AddPrototypes(protos ...desc.Data) error {
	for _, proto := range protos {
		mod.prototypes.Set(proto.Type(), proto)
	}
	return nil
}

func (mod *Module) UnmarshalDescriptor(name string, buf []byte) desc.Data {
	p, ok := mod.prototypes.Get(name)
	if !ok {
		return nil
	}
	var v = reflect.ValueOf(p)

	c := reflect.New(v.Type())

	err := json.Unmarshal(buf, c.Interface())
	if err != nil {
		panic(err)
	}

	return c.Elem().Interface().(desc.Data)
}

func (mod *Module) describe(_ context.Context, dataID data.ID, _ *desc.Opts) []*desc.Desc {
	var descs []*desc.Desc
	var err error
	var row dbDataType

	err = mod.db.Where("data_id = ?", dataID).First(&row).Error
	if err == nil {
		descs = append(descs, &desc.Desc{
			Source: mod.node.Identity(),
			Data: content.TypeDesc{
				Method:      row.Method,
				ContentType: row.Type,
			},
		})
	}

	if label := mod.GetLabel(dataID); label != "" {
		descs = append(descs, &desc.Desc{
			Source: mod.node.Identity(),
			Data:   content.LabelDesc{Label: label},
		})
	}

	return descs
}
