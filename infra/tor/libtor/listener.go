package libtor

import (
	bine "github.com/cretz/bine/tor"
	"github.com/cretz/bine/torutil/ed25519"
	"github.com/cryptopunkscc/astrald/infra/tor"
	"net"
)

var _ tor.Listener = &listener{}

type listener struct {
	service *bine.OnionService
}

func (l listener) Accept() (net.Conn, error) {
	return l.service.Accept()
}

func (l listener) Close() error {
	return l.service.Close()
}

func (l listener) Addr() string {
	return l.service.Addr().String()
}

func (l listener) ed25519Key() ed25519.KeyPair {
	return l.service.Key.(ed25519.KeyPair)
}

func (l listener) PrivateKey() tor.Key {
	return tor.Key(l.ed25519Key().PrivateKey())
}
