package keys

import (
	"github.com/cryptopunkscc/astrald/astral"
)

type IndexerService struct {
	*Module
}

func (srv *IndexerService) Run(ctx *astral.Context) error {
	<-ctx.Done()

	return nil
}
