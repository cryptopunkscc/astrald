package astral

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/mod/apphost/ipc"
	"github.com/cryptopunkscc/astrald/proto/apphost"
	"io"
)

var ListenProtocol = "unix"
var AppHostAddress string

func Connect() (apphost.AppHost, error) {
	conn, err := ipc.Dial(AppHostAddress)
	return apphost.Bind(conn), err
}

func Listen(port string) (*Listener, error) {
	l, err := NewListener("unix")
	if err != nil {
		return nil, err
	}

	host, err := Connect()
	if err != nil {
		return nil, err
	}

	err = host.Register(port, l.Target())
	if err != nil {
		l.Close()
		return nil, err
	}

	return l, nil
}

func Dial(remoteID id.Identity, query string) (io.ReadWriteCloser, error) {
	host, err := Connect()
	if err != nil {
		return nil, err
	}

	return host.Query(remoteID, query)
}

func DialName(nodeName string, query string) (io.ReadWriteCloser, error) {
	identity, err := Resolve(nodeName)
	if err != nil {
		return nil, err
	}

	return Dial(identity, query)
}

func Resolve(s string) (id.Identity, error) {
	host, err := Connect()
	if err != nil {
		return id.Identity{}, err
	}

	return host.Resolve(s)
}

func NodeInfo(identity id.Identity) (apphost.NodeInfo, error) {
	host, err := Connect()
	if err != nil {
		return apphost.NodeInfo{}, err
	}

	return host.NodeInfo(identity)
}
