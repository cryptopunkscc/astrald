package zip

import (
	"context"
	"github.com/cryptopunkscc/astrald/mod/content"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/mod/sets"
	"github.com/cryptopunkscc/astrald/node/events"
	"github.com/cryptopunkscc/astrald/object"
)

const zipMimeType = "application/zip"

type IndexerService struct {
	*Module
}

func (srv *IndexerService) Run(ctx context.Context) error {
	go events.Handle(ctx, srv.node.Events(), func(event sets.EventMemberUpdate) error {
		if event.Removed {
			srv.onRemove(event.ObjectID)
		} else {
			srv.onAdd(event.ObjectID)
		}
		return nil
	})

	for event := range srv.content.Scan(ctx, &content.ScanOpts{Type: zipMimeType}) {
		srv.autoIndexZip(event.ObjectID)
	}

	<-ctx.Done()

	return nil
}

func (srv *IndexerService) onAdd(objectID object.ID) error {
	if !srv.isIndexed(objectID, false) {
		found, err := srv.objects.Open(
			objectID,
			&objects.OpenOpts{
				Virtual: srv.config.Virtual,
				Network: srv.config.Network,
			},
		)
		if err != nil {
			return nil
		}
		found.Close()
		srv.Index(objectID)
	}
	return nil
}

func (srv *IndexerService) onRemove(objectID object.ID) error {
	if srv.isIndexed(objectID, false) {
		found, err := srv.objects.Open(
			objectID,
			&objects.OpenOpts{
				Virtual: srv.config.Virtual,
				Network: srv.config.Network,
			},
		)
		if err != nil {
			srv.Unindex(objectID)
			return nil
		}
		found.Close()
	}
	return nil
}

func (srv *IndexerService) autoIndexZip(zipID object.ID) error {
	// check if the file is accessible
	found, err := srv.objects.Open(
		zipID,
		&objects.OpenOpts{
			Virtual: srv.config.Virtual,
			Network: srv.config.Network,
		},
	)
	if err != nil {
		return err
	}
	found.Close()

	return srv.Index(zipID)
}
