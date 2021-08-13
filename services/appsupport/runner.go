package appsupport

import (
	"context"
	"github.com/cryptopunkscc/astrald/api"
	"github.com/cryptopunkscc/astrald/node"
)

func init() {
	_ = node.RegisterService("apps", func(ctx context.Context, core api.Core) error {
		return (&AppSupport{
			network: core.Network(),
		}).Run(ctx)
	})
}
