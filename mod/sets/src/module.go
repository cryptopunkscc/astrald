package sets

import (
	"context"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/core/assets"
	"github.com/cryptopunkscc/astrald/events"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/sets"
	"github.com/cryptopunkscc/astrald/object"
	"gorm.io/gorm"
)

var _ sets.Module = &Module{}

type Module struct {
	config Config
	node   astral.Node
	log    *log.Logger
	assets assets.Assets
	events events.Queue
	db     *gorm.DB
}

func (mod *Module) Run(ctx context.Context) error {
	<-ctx.Done()

	return nil
}

func (mod *Module) Create(name string) (sets.Set, error) {
	var row = dbSet{
		Name: name,
	}

	err := mod.db.Create(&row).Error
	if err != nil {
		return nil, err
	}

	var set = &Set{
		Module: mod,
		row:    &row,
	}

	mod.events.Emit(sets.EventSetCreated{Set: set})

	return set, nil
}

func (mod *Module) Open(name string, create bool) (sets.Set, error) {
	var row dbSet
	var err = mod.db.Where("name = ?", name).First(&row).Error
	if err != nil {
		if create {
			return mod.Create(name)
		}
		return nil, err
	}

	var set = &Set{
		Module: mod,
		row:    &row,
	}

	return set, nil
}

func (mod *Module) Where(objectID object.ID) ([]string, error) {
	dataRow, err := mod.dbDataFindOrCreateByObjectID(objectID)
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
