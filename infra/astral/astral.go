package astral

import (
	"context"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/enc"
	"github.com/cryptopunkscc/astrald/infra"
	"io"
	"log"
	"strings"
)

const NetworkName = "astral"

var _ infra.Network = &Astral{}

type Node interface {
	Identity() id.Identity
	Query(ctx context.Context, remoteID id.Identity, query string) (io.ReadWriteCloser, error)
}

type Astral struct {
	Node
	config      Config
	publicAddrs []infra.AddrDesc
}

func NewAstral(node Node, config Config) *Astral {
	a := &Astral{
		Node:        node,
		config:      config,
		publicAddrs: make([]infra.AddrDesc, 0),
	}

	// Add public addresses
	for _, gateID := range config.Gateways {
		gate, err := id.ParsePublicKeyHex(gateID)
		if err != nil {
			log.Println("astral: parse error:", err)
			continue
		}

		a.publicAddrs = append(a.publicAddrs, infra.AddrDesc{
			Addr: Addr{
				gate:   gate,
				target: node.Identity().Public(),
			},
			Public: false,
		})
		log.Println("astral: added", gate)
	}

	return a
}

func Parse(str string) (Addr, error) {
	parts := strings.Split(str, ":")
	if len(parts) != 2 {
		return Addr{}, infra.ErrInvalidAddress
	}

	targetKey := parts[1]
	gateKey := parts[0]

	gate, err := id.ParsePublicKeyHex(gateKey)
	if err != nil {
		return Addr{}, infra.ErrInvalidAddress
	}

	target, err := id.ParsePublicKeyHex(targetKey)
	if err != nil {
		return Addr{}, infra.ErrInvalidAddress
	}

	return Addr{
		gate:   gate,
		target: target,
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

	log.Println("connecting to", a.gate.PublicKeyHex())

	rwc, err := astral.Query(ctx, a.gate, ".gateway")
	if err != nil {
		log.Println("connect error:", err)
		return nil, err
	}

	log.Println("sending identity", a.target.PublicKeyHex())

	enc.WriteIdentity(rwc, a.target)

	res, err := enc.ReadUint8(rwc)
	if err != nil {
		log.Println("rejected", err)
		rwc.Close()
		return nil, infra.ErrConnectionRefused
	}

	if res != 1 {
		log.Println("rejected", res)
		rwc.Close()
		return nil, infra.ErrConnectionRefused
	}

	log.Println("connected", res)

	return newConn(rwc, a, true), nil
}

func (astral Astral) Addresses() []infra.AddrDesc {
	return astral.publicAddrs
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
