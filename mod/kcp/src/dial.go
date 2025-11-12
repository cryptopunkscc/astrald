package kcp

import (
	"fmt"
	"net"
	"time"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/exonet"
	"github.com/cryptopunkscc/astrald/mod/kcp"
	"github.com/pkg/errors"
	kcpgo "github.com/xtaci/kcp-go/v5"
)

var _ exonet.Dialer = &Module{}

// Dial establishes a KCP session and wraps it as an exonet.Conn.
func (mod *Module) Dial(ctx *astral.Context, endpoint exonet.Endpoint) (
	c exonet.Conn, err error) {
	if endpoint.Network() != "kcp" {
		return nil, exonet.ErrUnsupportedNetwork
	}

	remoteEndpoint, ok := endpoint.(*kcp.Endpoint)
	if !ok {
		return nil, fmt.Errorf("kcp/dial: endpoint is not a kcp endpoint")
	}

	udpConn, err := mod.prepareUDPConn(remoteEndpoint)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			_ = udpConn.Close()
		}
	}()

	kcpConn, err := kcpgo.NewConn(endpoint.Address(), nil, 0, 0, udpConn)
	if err != nil {
		return nil, fmt.Errorf("kcp/dial: creating KCP conn failed: %w", err)
	}

	defer func() {
		if err != nil {
			_ = kcpConn.Close()
		}
	}()

	// Without this deadline, dialing KCP can hang
	kcpConn.SetDeadline(time.Now().Add(mod.config.DialTimeout))

	localEndpoint, err := kcp.ParseEndpoint(kcpConn.LocalAddr().String())
	if err != nil {
		return nil, fmt.Errorf("kcp/dial: parsing local endpoint failed: %w", err)
	}

	return WrapKCPConn(kcpConn, remoteEndpoint, localEndpoint, true), nil
}

// prepareUDPConn creates a UDP connection, binding to an ephemeral local port if mapped.
func (mod *Module) prepareUDPConn(endpoint *kcp.Endpoint) (*net.UDPConn, error) {
	laddr := &net.UDPAddr{Port: 0}

	// Use mapped local port if available
	if port, ok := mod.ephemeralPortMappings.Get(endpoint.Address()); ok {
		laddr.Port = int(port)
	}

	conn, err := net.ListenUDP("udp", laddr)
	if err != nil {
		return nil, errors.Wrap(err, "kcp/dial: listening on ephemeral port failed")
	}

	return conn, nil
}
