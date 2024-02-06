package sets

import (
	"context"
	"github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/sets"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/node/assets"
	"github.com/cryptopunkscc/astrald/node/events"
	"github.com/cryptopunkscc/astrald/sig"
	"github.com/cryptopunkscc/astrald/tasks"
	"gorm.io/gorm"
)

var _ sets.Module = &Module{}

type Module struct {
	config   Config
	node     node.Node
	log      *log.Logger
	assets   assets.Assets
	events   events.Queue
	db       *gorm.DB
	wrappers sig.Map[string, sets.WrapperFunc]

	universe sets.Union
	device   sets.Union
	virtual  sets.Union
	network  sets.Union
}

func (mod *Module) Run(ctx context.Context) error {
	return tasks.Group(
		tasks.RunFuncAdapter{RunFunc: mod.watchUnionMembers},
	).Run(ctx)
}

func (mod *Module) SetWrapper(typ sets.Type, manager sets.WrapperFunc) {
	mod.wrappers.Replace(string(typ), manager)
}

func (mod *Module) Wrapper(typ sets.Type) sets.WrapperFunc {
	v, _ := mod.wrappers.Get(string(typ))
	return v
}

func (mod *Module) Stat(name string) (*sets.Stat, error) {
	setRow, err := mod.dbFindSetByName(name)
	if err != nil {
		return nil, err
	}

	var info = &sets.Stat{
		Name:        setRow.Name,
		Type:        sets.Type(setRow.Type),
		Size:        -1,
		Visible:     setRow.Visible,
		Description: setRow.Description,
		CreatedAt:   setRow.CreatedAt,
		TrimmedAt:   setRow.TrimmedAt,
	}

	var rows []dbMember

	err = mod.db.
		Model(&dbMember{}).
		Preload("Data").
		Where("set_id = ? and removed = false", setRow.ID).
		Find(&rows).Error
	if err != nil {
		return nil, err
	}

	for _, row := range rows {
		info.DataSize += row.Data.DataID.Size
	}

	info.Size = len(rows)

	return info, nil
}

func (mod *Module) Where(dataID data.ID) ([]string, error) {
	dataRow, err := mod.dbDataFindOrCreateByDataID(dataID)
	if err != nil {
		return nil, err
	}

	var rows []dbMember
	err = mod.db.
		Preload("Set").
		Where("data_id = ?", dataRow.ID).
		Find(&rows).Error

	var list []string

	for _, row := range rows {
		if !row.Removed {
			list = append(list, row.Set.Name)
		}
	}

	return list, nil
}

func (mod *Module) All() ([]string, error) {
	var list []string

	var err = mod.db.
		Model(&dbSet{}).
		Select("name").
		Find(&list).Error
	if err != nil {
		mod.log.Errorv(2, "All(): %v", err)
		return nil, err
	}

	return list, err
}

func (mod *Module) Scan(name string, opts *sets.ScanOpts) ([]*sets.Member, error) {
	set, err := mod.Open(name, false)
	if err != nil {
		return nil, err
	}

	return set.Scan(opts)
}

func (mod *Module) SetVisible(name string, visible bool) error {
	return mod.db.
		Model(&dbSet{}).
		Where("name = ?", name).
		Update("visible", visible).
		Error
}

func (mod *Module) SetDescription(name string, desc string) error {
	return mod.db.
		Model(&dbSet{}).
		Where("name = ?", name).
		Update("description", desc).
		Error
}

func (mod *Module) Universe() sets.Union {
	return mod.universe
}

func (mod *Module) Device() sets.Union {
	return mod.device
}

func (mod *Module) Virtual() sets.Union {
	return mod.virtual
}

func (mod *Module) Network() sets.Union {
	return mod.network
}
