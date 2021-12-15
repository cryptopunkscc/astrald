package astral

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/enc"
	"github.com/cryptopunkscc/astrald/infra"
	"io"
	"log"
)

const NetworkName = "astral"

var _ infra.Network = &Astral{}

type Node interface {
	Identity() id.Identity
	Query(ctx context.Context, remoteID id.Identity, query string) (io.ReadWriteCloser, error)
}

type Astral struct {
	Node
	config   Config
	gateways []infra.AddrSpec
}

func NewAstral(node Node, config Config) *Astral {
	a := &Astral{
		Node:     node,
		config:   config,
		gateways: make([]infra.AddrSpec, 0),
	}

	// Add public addresses
	for _, gateID := range config.Gateways {
		gate, err := id.ParsePublicKeyHex(gateID)
		if err != nil {
			log.Println("astral: parse error:", err)
			continue
		}

		a.gateways = append(a.gateways, infra.AddrSpec{
			Addr: Addr{
				gate:   gate,
				target: node.Identity().Public(),
			},
			Global: true,
		})
	}

	return a
}

func Parse(str string) (Addr, error) {
	gate, err := id.ParsePublicKeyHex(str)
	if err != nil {
		return Addr{}, err
	}

	return Addr{
		gate: gate,
	}, nil
}

func (astral Astral) Name() string {
	return NetworkName
}

func (astral Astral) Unpack(bytes []byte) (infra.Addr, error) {
	return Unpack(bytes)
}

func (astral Astral) Dial(ctx context.Context, addr infra.Addr) (infra.Conn, error) {
	a, ok := addr.(Addr)
	if !ok {
		return nil, infra.ErrUnsupportedAddress
	}

	if a.gate.IsEqual(astral.Node.Identity()) {
		return nil, errors.New("cannot connect to self")
	}

	rwc, err := astral.Query(ctx, a.gate, ".gateway")
	if err != nil {
		return nil, err
	}

	enc.WriteIdentity(rwc, a.target)

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

func (astral Astral) Addresses() []infra.AddrSpec {
	return astral.gateways
}

func (astral Astral) Listen(context.Context) (<-chan infra.Conn, <-chan error) {
	return nil, singleErrCh(infra.ErrUnsupportedOperation)
}

func (astral Astral) Broadcast([]byte) error {
	return infra.ErrUnsupportedOperation
}

func (astral Astral) Scan(context.Context) (<-chan infra.Broadcast, <-chan error) {
	return nil, singleErrCh(infra.ErrUnsupportedOperation)
}

func (astral Astral) Announce(context.Context, id.Identity) error {
	return infra.ErrUnsupportedOperation
}

func (astral Astral) Discover(context.Context) (<-chan infra.Presence, error) {
	return nil, infra.ErrUnsupportedOperation
}

// singleErrCh creates a closed error channel with a single error in it
func singleErrCh(err error) <-chan error {
	ch := make(chan error, 1)
	defer close(ch)
	ch <- err
	return ch
}
