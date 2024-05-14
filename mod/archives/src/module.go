package archives

import (
	_zip "archive/zip"
	"context"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/archives"
	"github.com/cryptopunkscc/astrald/mod/content"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/mod/shares"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/node/events"
	"github.com/cryptopunkscc/astrald/object"
	"gorm.io/gorm"
	"sync"
)

const zipMimeType = "application/zip"

var _ archives.Module = &Module{}

type Module struct {
	config Config
	node   node.Node
	events events.Queue
	log    *log.Logger

	db      *gorm.DB
	content content.Module
	objects objects.Module
	shares  shares.Module

	mu            sync.Mutex
	autoIndexZone net.Zone
}

func (mod *Module) Run(ctx context.Context) error {
	mod.autoIndexZone = net.Zones(mod.config.AutoIndexZones)

	go events.Handle(ctx, mod.node.Events(), func(event objects.EventObjectDiscovered) error {
		return mod.onObjectDiscovered(ctx, event)
	})

	for event := range mod.content.Scan(ctx, &content.ScanOpts{Type: zipMimeType}) {
		mod.Index(ctx, event.ObjectID, &objects.OpenOpts{Zone: mod.autoIndexZone})
	}

	return nil
}

func (mod *Module) onObjectDiscovered(ctx context.Context, event objects.EventObjectDiscovered) error {
	info, _ := mod.content.Identify(event.ObjectID)
	if info != nil && info.Type == zipMimeType {
		archive, _ := mod.Index(
			ctx,
			event.ObjectID,
			&objects.OpenOpts{Zone: mod.autoIndexZone},
		)

		if archive == nil {
			return nil
		}

		for _, entry := range archive.Entries {
			mod.events.Emit(objects.EventObjectDiscovered{
				ObjectID: entry.ObjectID,
				Zone:     net.ZoneVirtual | event.Zone,
			})
		}
	}
	return nil
}

func (mod *Module) Open(ctx context.Context, objectID object.ID, opts *objects.OpenOpts) (objects.Reader, error) {
	if opts == nil {
		opts = &objects.OpenOpts{}
	}

	if !opts.Zone.Is(net.ZoneVirtual) {
		return nil, net.ErrZoneExcluded
	}

	if opts.Offset > objectID.Size {
		return nil, objects.ErrInvalidOffset
	}

	var rows []dbEntry
	err := mod.db.
		Unscoped().
		Preload("Parent").
		Where("object_id = ?", objectID).
		Find(&rows).Error
	if err != nil {
		return nil, err
	}

	for _, row := range rows {
		r, err := mod.open(row.Parent.ObjectID, row.Path, row.ObjectID, opts)
		if err == nil {
			mod.log.Logv(2, "opened %v from %v/%v", objectID, row.Parent.ObjectID, row.Path)
			return r, nil
		}
	}

	return nil, objects.ErrNotFound
}

func (mod *Module) open(zipID object.ID, path string, fileID object.ID, opts *objects.OpenOpts) (objects.Reader, error) {
	zipFile, err := mod.openZip(zipID, opts)
	if err != nil {
		return nil, objects.ErrNotFound
	}

	var r = &contentReader{
		zip:      zipFile,
		path:     path,
		objectID: fileID,
	}

	err = r.open()

	return r, err
}

func (mod *Module) openZip(objectID object.ID, opts *objects.OpenOpts) (*_zip.Reader, error) {
	var r = &readerAt{
		objects:  mod.objects,
		objectID: objectID,
		opts:     opts,
	}

	zipFile, err := _zip.NewReader(r, int64(objectID.Size))
	return zipFile, err
}
