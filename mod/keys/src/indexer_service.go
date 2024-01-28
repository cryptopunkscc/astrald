package keys

import (
	"context"
	"github.com/cryptopunkscc/astrald/mod/content"
	"github.com/cryptopunkscc/astrald/mod/keys"
)

type IndexerService struct {
	*Module
}

func (srv *IndexerService) Run(ctx context.Context) error {
	for event := range srv.content.Scan(ctx, &content.ScanOpts{Type: keys.PrivateKeyDataType}) {
		srv.IndexKey(event.DataID)
	}

	<-ctx.Done()

	return nil
}
