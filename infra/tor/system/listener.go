package system

import (
	"encoding/base64"
	"github.com/cryptopunkscc/astrald/infra/tor"
	"github.com/cryptopunkscc/astrald/infra/tor/tc"
	"net"
	"strings"
)

var _ tor.Listener = &listener{}

type listener struct {
	tcp   net.Listener
	onion tc.Onion
	ctl   *tc.Control
}

func (l listener) Addr() string {
	return l.onion.ServiceID
}

func (l listener) PrivateKey() tor.Key {
	s := l.onion.PrivateKey

	// force v3 as v2 is now considered insecure
	if !strings.HasPrefix(s, "ED25519-V3:") {
		return nil
	}

	ss := strings.Split(s, ":")

	if len(ss) != 2 {
		return nil
	}

	key, err := base64.StdEncoding.DecodeString(ss[1])
	if err != nil {
		return nil
	}

	return key
}

func (l listener) Accept() (net.Conn, error) {
	return l.tcp.Accept()
}

func (l listener) Close() error {
	l.tcp.Close()

	return l.ctl.DelOnion(l.onion.ServiceID)
}
