package apphost

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/lib/astral"
	"github.com/cryptopunkscc/astrald/lib/wrapper"
	"io"
)

var _ wrapper.Api = Adapter{}

type Adapter struct{}

func (a Adapter) Resolve(name string) (id.Identity, error) {
	return astral.Resolve(name)
}

func (a Adapter) Register(name string) (wrapper.Port, error) {
	listener, err := astral.Listen(name)
	if err != nil {
		return nil, err
	}
	return appHostPort{listener}, err
}

func (a Adapter) Query(nodeID id.Identity, query string) (rw io.ReadWriteCloser, err error) {
	return astral.Dial(nodeID, query)
}

type appHostPort struct{ *astral.Listener }

func (a appHostPort) Next() <-chan wrapper.Request {
	c := make(chan wrapper.Request)
	go func() {
		defer close(c)
		for query := range a.QueryCh() {
			q := query
			c <- &appHostRequest{q}
		}
	}()
	return c
}

func (a appHostPort) Close() error {
	return a.Listener.Close()
}

type appHostRequest struct{ query *astral.Query }

func (a appHostRequest) Caller() id.Identity {
	return a.query.RemoteIdentity()
}

func (a appHostRequest) Accept() (io.ReadWriteCloser, error) {
	return a.query.Accept()
}

func (a appHostRequest) Reject() error {
	return a.query.Reject()
}

func (a appHostRequest) Query() string {
	return a.query.Query()
}
