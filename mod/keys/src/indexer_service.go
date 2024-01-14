package keys

import (
	"context"
	"github.com/cryptopunkscc/astrald/mod/keys"
	"time"
)

type IndexerService struct {
	*Module
}

func (srv *IndexerService) Run(ctx context.Context) error {
	for event := range srv.data.SubscribeType(ctx, keys.PrivateKeyDataType, time.Time{}) {
		srv.IndexKey(event.DataID)
	}

	<-ctx.Done()

	return nil
}
