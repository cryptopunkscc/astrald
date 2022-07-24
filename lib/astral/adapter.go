package astral

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"io"
)

var adapter Api = appHostAdapter{}

func AppHostAdapter() Api {
	return adapter
}

type appHostAdapter struct{}

func (a appHostAdapter) Resolve(name string) (id.Identity, error) {
	return Resolve(name)
}

func (a appHostAdapter) Register(name string) (Port, error) {
	listener, err := Listen(name)
	if err != nil {
		return nil, err
	}
	return appHostPort{listener}, err
}

func (a appHostAdapter) Query(nodeID id.Identity, query string) (rw io.ReadWriteCloser, err error) {
	return Dial(nodeID, query)
}

type appHostPort struct{ *Listener }

func (a appHostPort) Next() <-chan Request {
	c := make(chan Request)
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
	return a.listener.Close()
}

type appHostRequest struct{ query *Query }

func (a appHostRequest) Caller() id.Identity {
	return a.query.remoteID
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
