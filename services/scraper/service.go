package scraper

import (
	"context"
	"github.com/cryptopunkscc/astrald/api"
	"github.com/cryptopunkscc/astrald/services/util/accept"
	"github.com/cryptopunkscc/astrald/services/util/register"
)

func run(ctx context.Context, core api.Core) error {
	handler, err := register.Port(ctx, core, Port)
	if err != nil {
		return err
	}

	for request := range handler.Requests() {
		_ = accept.Request(ctx, request)
	}

	return nil
}
