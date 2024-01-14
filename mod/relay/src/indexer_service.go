package relay

import (
	"context"
	"github.com/cryptopunkscc/astrald/mod/relay"
	"time"
)

type IndexerService struct {
	*Module
}

func (srv *IndexerService) Run(ctx context.Context) error {
	for event := range srv.data.SubscribeType(ctx, "", time.Time{}) {
		switch event.Type {
		case relay.RelayCertType:
			err := srv.IndexCert(event.DataID)
			switch err {
			case nil:
				srv.log.Infov(1, "added certificate %v", event.DataID)
			case relay.ErrCertAlreadyIndexed:
				// nothing
			default:
				srv.log.Errorv(1, "error adding cert %v: %v", event.DataID, err)
			}
		}
	}

	<-ctx.Done()

	return nil
}
