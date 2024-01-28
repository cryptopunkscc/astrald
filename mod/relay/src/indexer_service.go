package relay

import (
	"context"
	"github.com/cryptopunkscc/astrald/mod/relay"
	"github.com/cryptopunkscc/astrald/mod/storage"
)

type IndexerService struct {
	*Module
}

func (srv *IndexerService) Run(ctx context.Context) error {
	for event := range srv.content.Scan(ctx, nil) {
		switch event.Type {
		case relay.RelayCertType:
			err := srv.IndexCert(event.DataID)
			switch err {
			case nil:
				srv.log.Infov(1, "added certificate %v", event.DataID)
			case relay.ErrCertAlreadyIndexed, storage.ErrNotFound:
				// nothing
			default:
				srv.log.Errorv(1, "error adding cert %v: %v", event.DataID, err)
			}
		}
	}

	<-ctx.Done()

	return nil
}
