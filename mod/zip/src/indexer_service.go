package zip

import (
	"context"
	"github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/mod/content"
	"github.com/cryptopunkscc/astrald/mod/sets"
	"github.com/cryptopunkscc/astrald/mod/storage"
	"github.com/cryptopunkscc/astrald/node/events"
)

const zipMimeType = "application/zip"

type IndexerService struct {
	*Module
}

func (srv *IndexerService) Run(ctx context.Context) error {
	go events.Handle(ctx, srv.node.Events(), func(event sets.EventMemberUpdate) error {
		if event.Removed {
			srv.onRemove(event.DataID)
		} else {
			srv.onAdd(event.DataID)
		}
		return nil
	})

	for event := range srv.content.Scan(ctx, &content.ScanOpts{Type: zipMimeType}) {
		srv.autoIndexZip(event.DataID)
	}

	<-ctx.Done()

	return nil
}

func (srv *IndexerService) onAdd(dataID data.ID) error {
	if !srv.isIndexed(dataID, false) {
		found, err := srv.storage.Read(
			dataID,
			&storage.ReadOpts{
				Virtual: srv.config.Virtual,
				Network: srv.config.Network,
			},
		)
		if err != nil {
			return nil
		}
		found.Close()
		srv.Index(dataID)
	}
	return nil
}

func (srv *IndexerService) onRemove(dataID data.ID) error {
	if srv.isIndexed(dataID, false) {
		found, err := srv.storage.Read(
			dataID,
			&storage.ReadOpts{
				Virtual: srv.config.Virtual,
				Network: srv.config.Network,
			},
		)
		if err != nil {
			srv.Unindex(dataID)
			return nil
		}
		found.Close()
	}
	return nil
}

func (srv *IndexerService) autoIndexZip(zipID data.ID) error {
	// check if the file is accessible
	found, err := srv.storage.Read(
		zipID,
		&storage.ReadOpts{
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
