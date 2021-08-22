package share

import (
	"context"
	"github.com/cryptopunkscc/astrald/api"
	"github.com/cryptopunkscc/astrald/components/shares"
	"github.com/cryptopunkscc/astrald/services/util/auth"
	"github.com/cryptopunkscc/astrald/services/util/handle"
	"github.com/cryptopunkscc/astrald/services/util/request"
)

const Port = "share"

const (
	Add      = 1
	Remove   = 2
	List     = 3
	Contains = 4
)

func runService(ctx context.Context, core api.Core) error {
	rc := NewContext(shares.NewMemShares())
	handlers := map[byte]request.Handler{
		Add:      rc.Add,
		Remove:   rc.Remove,
		List:     rc.List,
		Contains: rc.Contains,
	}
	handle.Requests(ctx, core, Port, auth.Local, handle.Using(handlers))
	return nil
}
