package media

import (
	"context"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/lib/desc"
	"slices"
)

type IndexerService struct {
	*Module
}

func (srv *IndexerService) Run(ctx context.Context) error {
	for event := range srv.content.Scan(ctx, nil) {
		opts := &desc.Opts{}
		if slices.Contains(srv.config.AutoIndexNet, event.Type) {
			opts.Network = true
			opts.IdentityFilter = id.AllowEveryone
		}

		srv.Describe(ctx, event.DataID, opts)
	}

	return nil
}
