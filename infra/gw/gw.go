package gw

import (
	"context"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/infra"
	"io"
)

const NetworkName = "gw"
const PortName = "gateway"

var _ infra.Network = &Gateway{}

type Querier interface {
	Query(ctx context.Context, remoteID id.Identity, query string) (io.ReadWriteCloser, error)
}

type Gateway struct {
	Querier
	config Config
}

func New(config Config, querier Querier) (*Gateway, error) {
	return &Gateway{
		Querier: querier,
		config:  config,
	}, nil
}

func (g *Gateway) Run(ctx context.Context) error {
	<-ctx.Done()
	return nil
}
