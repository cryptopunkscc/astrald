package zip

import (
	"context"
	"fmt"
	"github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/mod/storage"
	"time"
)

const zipMimeType = "application/zip"

type IndexerService struct {
	*Module
}

func (srv *IndexerService) Run(ctx context.Context) error {
	for event := range srv.data.SubscribeType(ctx, zipMimeType, time.Time{}) {
		fmt.Println(event.Type)
		srv.autoIndexZip(event.DataID)
	}

	<-ctx.Done()

	return nil
}

func (srv *IndexerService) autoIndexZip(zipID data.ID) error {
	// check if the file is accessible
	found, err := srv.storage.Data().Read(
		zipID,
		&storage.ReadOpts{NoVirtual: srv.config.NoVirtual},
	)
	if err != nil {
		return err
	}
	found.Close()

	return srv.indexZip(zipID, false)
}
