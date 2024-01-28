package zip

import (
	"context"
	"github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/mod/storage"
)

const zipMimeType = "application/zip"

type IndexerService struct {
	*Module
}

func (srv *IndexerService) Run(ctx context.Context) error {
	for event := range srv.content.Scan(ctx, nil) {
		srv.autoIndexZip(event.DataID)
	}

	<-ctx.Done()

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

	return srv.Index(zipID, false)
}
