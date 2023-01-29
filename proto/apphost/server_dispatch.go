package apphost

import (
	"errors"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/cslq/rpc"
	"github.com/cryptopunkscc/astrald/hub"
	"github.com/cryptopunkscc/astrald/streams"
	"io"
	"strings"
)

type serverDispatcher struct {
	conn   io.ReadWriteCloser
	server AppHost
}

func Serve(conn io.ReadWriteCloser, server AppHost) (err error) {
	defer conn.Close()

	var dispatcher = &serverDispatcher{conn: conn, server: server}

	err = rpc.Dispatch(dispatcher.conn, "[c]c", dispatcher.dispatch)

	if err != nil {
		if errors.Is(err, io.EOF) || strings.Contains(err.Error(), "use of closed network connection") {
			err = nil
		}
	}
	return
}

func (dispatcher *serverDispatcher) dispatch(cmd string) error {
	switch cmd {
	case cmdRegister:
		return rpc.Dispatch(dispatcher.conn, "[c]c [c]c", dispatcher.register)

	case cmdQuery:
		return rpc.Dispatch(dispatcher.conn, "v [c]c", dispatcher.query)

	case cmdResolve:
		return rpc.Dispatch(dispatcher.conn, "[c]c", dispatcher.resolve)

	case cmdNodeInfo:
		return rpc.Dispatch(dispatcher.conn, "v", dispatcher.nodeInfo)

	default:
		return ErrUnknownCommand
	}
}

func (dispatcher *serverDispatcher) register(portName string, target string) error {
	err := dispatcher.server.Register(portName, target)

	switch {
	case err == nil:
		return cslq.Encode(dispatcher.conn, "c", success)

	case errors.Is(err, hub.ErrAlreadyRegistered), errors.Is(err, ErrAlreadyRegistered):
		return cslq.Encode(dispatcher.conn, "c", errAlreadyRegistered)

	default:
		return cslq.Encode(dispatcher.conn, "c", errFailed)
	}
}

func (dispatcher *serverDispatcher) query(identity id.Identity, query string) error {
	conn, err := dispatcher.server.Query(identity, query)
	switch {
	case err == nil:
		if err := cslq.Encode(dispatcher.conn, "c", success); err != nil {
			return err
		}

	case errors.Is(err, ErrRejected):
		return cslq.Encode(dispatcher.conn, "c", errRejected)

	case errors.Is(err, ErrTimeout):
		return cslq.Encode(dispatcher.conn, "c", errTimeout)

	default:
		return cslq.Encode(dispatcher.conn, "c", errUnexpected)
	}

	_, _, err = streams.Join(dispatcher.conn, conn)

	return err
}

func (dispatcher *serverDispatcher) resolve(s string) error {
	identity, err := dispatcher.server.Resolve(s)
	if err != nil {
		return cslq.Encode(dispatcher.conn, "c", errFailed)
	}

	return cslq.Encode(dispatcher.conn, "c v", success, identity)
}

func (dispatcher *serverDispatcher) nodeInfo(identity id.Identity) error {
	var code int

	info, err := dispatcher.server.NodeInfo(identity)
	if err != nil {
		code = errFailed
	}

	return cslq.Encode(dispatcher.conn, "c v", code, info)
}
