package keys

import (
	"errors"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/content"
	"github.com/cryptopunkscc/astrald/mod/keys"
)

type IndexerService struct {
	*Module
}

func (srv *IndexerService) Run(ctx *astral.Context) error {
	for event := range srv.Content.Scan(ctx, &content.ScanOpts{Type: keys.PrivateKey{}.ObjectType()}) {
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
