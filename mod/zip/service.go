package zip

import (
	"context"
	"github.com/cryptopunkscc/astrald/data"
	_data "github.com/cryptopunkscc/astrald/mod/data/api"
	storage "github.com/cryptopunkscc/astrald/mod/storage/api"
	"time"
)

const zipMimeType = "application/zip"

type Service struct {
	*Module
}

func (srv *Service) Run(ctx context.Context) error {
	go srv.handleEvents(ctx)

	list, err := srv.data.FindByType(zipMimeType, time.Time{})
	if err == nil {
		for _, item := range list {
			srv.autoIndexZip(item.ID)
		}
	} else {
		srv.log.Error("data.FindByType: %v", err)
	}

	<-ctx.Done()

	return nil
}

func (srv *Service) handleEvents(ctx context.Context) {
	for event := range srv.node.Events().Subscribe(ctx) {
		switch event := event.(type) {
		case _data.EventDataIndexed:
			if event.Type != zipMimeType {
				continue
			}

			srv.autoIndexZip(event.ID)
		}
	}
}

func (srv *Service) autoIndexZip(zipID data.ID) error {
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
