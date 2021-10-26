package linker

import (
	"context"
	"github.com/cryptopunkscc/astrald/astral/link"
)

type Strategy interface {
	Run(ctx context.Context) <-chan *link.Link
}
