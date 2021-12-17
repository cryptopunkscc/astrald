package gw

import (
	"context"
	"errors"
	"fmt"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/enc"
	"github.com/cryptopunkscc/astrald/infra"
	"io"
	"log"
	"strings"
)

const NetworkName = "gw"
const PortName = "gateway"

var _ infra.Network = &Gateway{}

type Querier interface {
	Query(ctx context.Context, remoteID id.Identity, query string) (io.ReadWriteCloser, error)
}

type Gateway struct {
	Querier
	config   Config
	gateways []infra.AddrSpec
}

func New(querier Querier, config Config) *Gateway {
	gw := &Gateway{
		Querier:  querier,
		config:   config,
		gateways: make([]infra.AddrSpec, 0),
	}

	// Add pre-configured gateways
	for _, gateStr := range config.Gateways {
		gate, err := Parse(gateStr)
		if err != nil {
			log.Println("config: invalid gateway:", err)
			continue
		}

		gw.gateways = append(gw.gateways, infra.AddrSpec{
			Addr:   gate,
			Global: true,
		})
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

func (Gateway) Unpack(bytes []byte) (infra.Addr, error) {
	return Unpack(bytes)
}

func (g Gateway) Dial(ctx context.Context, addr infra.Addr) (infra.Conn, error) {
	a, ok := addr.(Addr)
	if !ok {
		return nil, infra.ErrUnsupportedAddress
	}

	if len(a.cookie) == 0 {
		return nil, errors.New("missing cookie")
	}

	rwc, err := g.Query(ctx, a.gate, PortName)
	if err != nil {
		return nil, fmt.Errorf("gateway query error: %w", err)
	}

	if err := enc.WriteL8String(rwc, a.cookie); err != nil {
		return nil, fmt.Errorf("gateway query error: %w", err)
	}

	res, err := enc.ReadUint8(rwc)
	if err != nil {
		rwc.Close()
		return nil, infra.ErrConnectionRefused
	}

	if res != 1 {
		rwc.Close()
		return nil, infra.ErrConnectionRefused
	}

	return newConn(rwc, a, true), nil
}

func (g Gateway) Addresses() []infra.AddrSpec {
	return g.gateways
}

func (Gateway) Listen(context.Context) (<-chan infra.Conn, <-chan error) {
	return nil, singleErrCh(infra.ErrUnsupportedOperation)
}

func (Gateway) Broadcast([]byte) error {
	return infra.ErrUnsupportedOperation
}

func (Gateway) Scan(context.Context) (<-chan infra.Broadcast, <-chan error) {
	return nil, singleErrCh(infra.ErrUnsupportedOperation)
}

func (Gateway) Announce(context.Context, id.Identity) error {
	return infra.ErrUnsupportedOperation
}

func (Gateway) Discover(context.Context) (<-chan infra.Presence, error) {
	return nil, infra.ErrUnsupportedOperation
}

// singleErrCh creates a closed error channel with a single error in it
func singleErrCh(err error) <-chan error {
	ch := make(chan error, 1)
	defer close(ch)
	ch <- err
	return ch
}
