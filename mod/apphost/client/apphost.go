package astral

import (
	"errors"
	"fmt"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/mod/apphost/ipc"
	"github.com/cryptopunkscc/astrald/mod/apphost/proto"
	"io"
)

var ListenProtocol = "unix"
var instance *AppHost

type AppHost struct {
	address string
}

func NewAppHost(address string) *AppHost {
	return &AppHost{address: address}
}

func (host *AppHost) Query(nodeName string, query string) (io.ReadWriteCloser, error) {
	conn, err := ipc.Dial(host.address)
	if err != nil {
		return nil, err
	}

	stream := cslq.NewEndec(conn)

	stream.Encode("c [c]c [c]c", proto.RequestDialString, nodeName, query)

	var result int

	stream.Decode("c", &result)

	if result != proto.ResponseOK {
		return nil, fmt.Errorf("error code %d", result)
	}

	return conn, nil
}

func (host *AppHost) Register(portName string) (*Listener, error) {
	conn, err := ipc.Dial(host.address)
	if err != nil {
		return nil, err
	}

	l, err := NewListener(ListenProtocol)
	if err != nil {
		conn.Close()
		return nil, err
	}

	stream := cslq.NewEndec(conn)

	stream.Encode("c [c]c [c]c", proto.RequestRegister, portName, l.Target())

	var result int

	stream.Decode("c", &result)

	if result != proto.ResponseOK {
		conn.Close()
		return nil, fmt.Errorf("error code %d", result)
	}

	l.portCloser = conn

	return l, nil
}

func (host *AppHost) GetNodeName(identity id.Identity) (string, error) {
	conn, err := ipc.Dial(host.address)
	if err != nil {
		return "", err
	}

	stream := cslq.NewEndec(conn)

	if err := stream.Encode("cv", proto.RequestGetNodeName, identity); err != nil {
		return "", err
	}

	var name string

	if err := stream.Decode("[c]c", &name); err != nil {
		return "", nil
	}

	return name, nil
}

func (host *AppHost) Identity() (id.Identity, error) {
	conn, err := ipc.Dial(host.address)
	if err != nil {
		return id.Identity{}, err
	}

	stream := cslq.NewEndec(conn)

	stream.Encode("c", proto.RequestInfo)

	var (
		result   int
		identity id.Identity
	)

	stream.Decode("c", &result)

	if result != proto.ResponseOK {
		return id.Identity{}, errors.New("error")
	}

	stream.Decode("v", &identity)

	return identity, nil
}

func Query(node string, query string) (io.ReadWriteCloser, error) {
	return instance.Query(node, query)
}

func Reqister(name string) (*Listener, error) {
	return instance.Register(name)
}

func GetNodeName(identity id.Identity) (string, error) {
	return instance.GetNodeName(identity)
}
