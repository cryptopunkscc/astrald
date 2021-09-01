package client

import (
	"errors"
	"fmt"
	"github.com/cryptopunkscc/astrald/api"
	"github.com/cryptopunkscc/astrald/services/apphost/proto"
	"log"
	"net"
)

// ===================== Constructor =====================

const tcpPort = 8625

func NewCoreAdapter() api.Core {
	return &coreAdapter{NewNetworkAdapter()}
}

func NewNetworkAdapter() api.Network {
	return &networkAdapter{
		network:     "tcp",
		appHostAddr: fmt.Sprintf("127.0.0.1:%d", tcpPort),
		localAddr:   "127.0.0.1:",
	}
}

// ===================== Core =====================

type coreAdapter struct {
	network api.Network
}

func (c *coreAdapter) Network() api.Network {
	return c.network
}

// ===================== Network =====================

type networkAdapter struct {
	network     string
	appHostAddr string
	localAddr   string
}

var _ api.Network = new(networkAdapter)

func (n *networkAdapter) Identity() api.Identity {
	return ""
}

func (n *networkAdapter) Register(name string) (api.PortHandler, error) {
	listener, err := net.Listen(n.network, n.localAddr)
	if err != nil {
		return nil, err
	}
	conn, err := net.Dial(n.network, n.appHostAddr)
	if err != nil {
		return nil, err
	}
	pc := proto.NewConn(conn)
	port := listener.Addr().(*net.TCPAddr).Port
	response, err := pc.Register(name, ":"+string(rune(port)))
	if err != nil {
		return nil, err
	}
	if response.Status != "ok" {
		return nil, errors.New(response.Error)
	}
	return &portHandlerAdapter{Listener: listener}, nil
}

func (n *networkAdapter) Connect(identity api.Identity, port string) (api.Stream, error) {
	conn, err := net.Dial(n.network, n.appHostAddr)
	if err != nil {
		return nil, err
	}
	pc := proto.NewConn(conn)
	response, err := pc.Connect(string(identity), port)
	if err != nil {
		return nil, err
	}
	if response.Status != "ok" {
		return nil, errors.New(response.Error)
	}
	return &streamAdapter{
		Conn: proto.NewConn(conn),
	}, nil
}

// ===================== Port Handler =====================

type portHandlerAdapter struct {
	port string
	net.Listener
}

var _ api.PortHandler = new(portHandlerAdapter)

func (p *portHandlerAdapter) Requests() <-chan api.ConnectionRequest {

	c := make(chan api.ConnectionRequest)
	go func() {
		for {
			conn, err := p.Listener.Accept()
			if err != nil {
				log.Println("cannot accept connection", err)
				continue
			}
			protoConn := proto.NewConn(conn)
			request, err := protoConn.ReadRequest()
			if err != nil {
				log.Println("cannot read connection request", err)
				continue
			}
			c <- &connectionRequestAdapter{
				Conn:    protoConn,
				request: request,
			}
		}
	}()
	return c
}

// ===================== Connection Request =====================

type connectionRequestAdapter struct {
	*proto.Conn
	request proto.Request
}

var _ api.ConnectionRequest = new(connectionRequestAdapter)

func (c *connectionRequestAdapter) Caller() api.Identity {
	return api.Identity(c.request.Identity)
}

func (c *connectionRequestAdapter) Query() string {
	return c.request.Port
}

func (c *connectionRequestAdapter) Accept() api.Stream {
	err := c.OK()
	if err != nil {
		log.Println("cannot accept", err)
		return nil
	}
	return streamAdapter{c.Conn}
}

func (c *connectionRequestAdapter) Reject() {
	err := c.Close()
	if err != nil {
		log.Println("rejection error", err)
	}
}

// ===================== Stream =====================

type streamAdapter struct {
	*proto.Conn
}

var _ api.Stream = new(streamAdapter)
