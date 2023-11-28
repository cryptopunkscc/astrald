package router

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/net"
)

const RerouteServiceName = "net.router.reroute"

type RerouteService struct {
	*Module
}

func (srv *RerouteService) Run(ctx context.Context) error {
	err := srv.node.AddRoute(RerouteServiceName, srv)
	if err != nil {
		return err
	}
	defer srv.node.RemoveRoute(RerouteServiceName)

	<-ctx.Done()

	return nil
}

func (srv *RerouteService) RouteQuery(ctx context.Context, query net.Query, caller net.SecureWriteCloser, hints net.Hints) (net.SecureWriteCloser, error) {
	return net.Accept(query, caller, func(conn net.SecureConn) {
		if err := srv.serve(conn); err != nil {
			srv.log.Errorv(1, "reroute serve: %v", err)
		}
	})
}

func (srv *RerouteService) serve(client net.SecureConn) error {
	var err error
	var nonce uint64

	err = cslq.Decode(client, "q", &nonce)
	if err != nil {
		return err
	}

	conn := srv.findConnByNonce(net.Nonce(nonce))
	if conn == nil {
		return errors.New("invalid nonce")
	}

	switcher, err := srv.insertSwitcherAfter(net.RootSource(conn.Target()))
	if err != nil {
		return err
	}

	newRoot, ok := net.RootSource(client).(net.OutputGetSetter)
	if !ok {
		return errors.New("newroot not an OutputGetSetter")
	}
	debris := newRoot.Output()
	newRoot.SetOutput(switcher.NextWriter)

	if err := cslq.Encode(client, "c", 0); err != nil {
		return err
	}

	oldOutput := net.FinalOutput(conn.Caller())
	newOutput := srv.yankFinalOutput(client)

	switcher.AfterSwitch = func() {
		if err := srv.replaceOutput(oldOutput, newOutput); err != nil {
			panic(err)
		}
		oldOutput.Close()
		debris.Close()
		client.Close()
	}

	return nil
}
