package gw

import (
	"context"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/infra"
	"io"
	"strings"
)

const NetworkName = "gw"
const PortName = "gateway"

var _ infra.Network = &Gateway{}
var _ infra.Dialer = &Gateway{}

type Querier interface {
	Query(ctx context.Context, remoteID id.Identity, query string) (io.ReadWriteCloser, error)
}

type Gateway struct {
	Querier
	config Config
}

func New(config Config, querier Querier) *Gateway {
	gw := &Gateway{
		Querier: querier,
		config:  config,
	}

	return gw
}

func Parse(str string) (Addr, error) {
	parts := strings.SplitN(str, ":", 2)

	gate, err := id.ParsePublicKeyHex(parts[0])
	if err != nil {
		return Addr{}, err
	}

	var cookie string
	if len(parts) == 2 {
		cookie = parts[1]
	}

	return Addr{
		gate:   gate,
		cookie: cookie,
	}, nil
}

func (Gateway) Name() string {
	return NetworkName
}
