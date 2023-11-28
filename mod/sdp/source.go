package sdp

import (
	"context"
	"github.com/cryptopunkscc/astrald/auth/id"
)

type Source interface {
	Discover(ctx context.Context, caller id.Identity, origin string) ([]ServiceEntry, error)
}
