package content

import (
	"context"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/events"
	"github.com/cryptopunkscc/astrald/lib/desc"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/content"
	"github.com/cryptopunkscc/astrald/mod/fs"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/object"
	"github.com/cryptopunkscc/astrald/sig"
	"gorm.io/gorm"
	"time"
)

var _ content.Module = &Module{}

const identifySize = 4096
const adcMethod = "adc"
const mimetypeMethod = "mimetype"

type Module struct {
	node   astral.Node
	config Config
	log    *log.Logger
	events events.Queue
	db     *gorm.DB

	prototypes sig.Map[string, desc.Data]
	ongoing    sig.Map[string, chan struct{}]
	objects    objects.Module
	fs         fs.Module

	ready chan struct{}
}

func (mod *Module) Run(ctx context.Context) error {
	go mod.identifyFS()

	go func() {
		for event := range mod.node.Events().Subscribe(ctx) {
			switch e := event.(type) {
			case objects.EventDiscovered:
				mod.Identify(e.ObjectID)
			}
		}
	}()

	<-ctx.Done()

	return nil
}

// Scan returns a channel that will be populated with all data entries since the provided timestamp and
// subscribed to any new items until context is done. If type is empty, all data entries will be passed regardless
// of the type.
func (mod *Module) Scan(ctx context.Context, opts *content.ScanOpts) <-chan *content.TypeInfo {
	if opts == nil {
		opts = &content.ScanOpts{}
	}

	if opts.After.After(time.Now()) {
		return nil
	}

	var ch = make(chan *content.TypeInfo)
	var subscription = mod.events.Subscribe(ctx)

	go func() {
		defer close(ch)

		// catch up with existing entries
		list, err := mod.scan(opts)
		if err != nil {
			return
		}
		for _, item := range list {
			select {
			case ch <- item:
			case <-ctx.Done():
				return
			}
		}

		// subscribe to new items
		for event := range subscription {
			e, ok := event.(content.EventObjectIdentified)
			if !ok {
				continue
			}
			if opts.Type != "" && e.TypeInfo.Type != opts.Type {
				continue
			}
			ch <- e.TypeInfo
		}
	}()

	return ch
}

func (mod *Module) Forget(objectID object.ID) error {
	return mod.db.Delete(&dbDataType{}, objectID).Error
}

func (mod *Module) Ready(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()

	case <-mod.ready:
		return nil
	}
}

// find returns all data items indexed since time ts. If t is not empty, only items of type t will
// be returned.
func (mod *Module) scan(opts *content.ScanOpts) ([]*content.TypeInfo, error) {
	var list []*content.TypeInfo
	var rows []*dbDataType

	if opts == nil {
		opts = &content.ScanOpts{}
	}

	var query = mod.db

	// filter by type if provided
	if opts.Type != "" {
		query = query.Where("type = ?", opts.Type)
	}

	// filter by time if provided
	if !opts.After.IsZero() {
		query = query.Where("identified_at > ?", opts.After)
	}

	// fetch rows
	var tx = query.Order("identified_at").Find(&rows)
	if tx.Error != nil {
		return nil, tx.Error
	}

	for _, row := range rows {
		list = append(list, &content.TypeInfo{
			ObjectID:     row.DataID,
			IdentifiedAt: row.IdentifiedAt,
			Method:       row.Method,
			Type:         row.Type,
		})
	}

	return list, nil
}

func (mod *Module) setReady() {
	close(mod.ready)
}
