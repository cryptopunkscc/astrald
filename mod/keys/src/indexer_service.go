package keys

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/mod/content"
	"github.com/cryptopunkscc/astrald/mod/keys"
)

type IndexerService struct {
	*Module
}

func (srv *IndexerService) Run(ctx context.Context) error {
	for event := range srv.content.Scan(ctx, &content.ScanOpts{Type: keys.PrivateKeyDataType}) {
		err := srv.IndexKey(event.ObjectID)
		switch {
		case err == nil:
		case errors.Is(err, ErrAlreadyIndexed):
		default:
			srv.log.Errorv(1, "IndexKey: %v", err)
		}
	}

	<-ctx.Done()

	return nil
}
