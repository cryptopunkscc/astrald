package kcp

import (
	"fmt"
	"time"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/exonet"
	kcpmod "github.com/cryptopunkscc/astrald/mod/kcp"
	kcpgo "github.com/xtaci/kcp-go/v5"
)

var _ exonet.Dialer = &Module{}

// Dial establishes a KCP session and wraps it as an exonet.Conn.
func (mod *Module) Dial(ctx *astral.Context, endpoint exonet.Endpoint) (
	c exonet.Conn, err error) {
	switch endpoint.Network() {
	case "kcp":
	default:
		return nil, exonet.ErrUnsupportedNetwork
	}

	sess, err := kcpgo.DialWithOptions(endpoint.Address(), nil, 0, 0)
	if err != nil {
		return nil, fmt.Errorf("kcp module/dial creating session failed: %w", err)
	}

	// NOTE: without this deadline, dialing KCP hangs (causes problems with re-attempts at establishing links)
	sess.SetDeadline(time.Now().Add(mod.config.DialTimeout))

	// Close raw session on any subsequent error; noop on success.
	defer func() {
		if err != nil {
			_ = sess.Close()
		}
	}()

	localEndpoint, err := kcpmod.ParseEndpoint(sess.LocalAddr().String())
	if err != nil {
		return nil, fmt.Errorf(`kcp module/dial parsing local endpoint failed: %w`, err)
	}

	remoteEndpoint, err := kcpmod.ParseEndpoint(sess.RemoteAddr().String())
	if err != nil {
		return nil, fmt.Errorf(`kcp module/dial parsing remote endpoint failed: %w`, err)
	}

	return WrapKCPConn(sess, remoteEndpoint, localEndpoint, true), nil
}
