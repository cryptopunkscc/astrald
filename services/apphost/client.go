package apphost

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/api"
	"github.com/cryptopunkscc/astrald/logfmt"
	"github.com/cryptopunkscc/astrald/services/apphost/proto"
	"log"
	"net"
)

type Client struct {
	socket *proto.Conn
	net    api.Network
}

func NewClient(conn net.Conn, netApi api.Network) *Client {
	return &Client{
		socket: proto.NewConn(conn),
		net:    netApi,
	}
}

func (c *Client) handle(ctx context.Context) error {
	request, err := c.socket.ReadRequest()
	if err != nil {
		return err
	}

	switch request.Type {
	case proto.RequestConnect:
		return c.handleConnect(ctx, request)

	case proto.RequestRegister:
		return c.handleRegister(ctx, request)

	default:
		_ = c.socket.Close() // ignore unknown requests
		return errors.New("invalid request")
	}
}

func (c *Client) handleConnect(ctx context.Context, request proto.Request) error {
	conn, err := c.net.Connect(api.Identity(request.Identity), request.Port)
	if err != nil {
		return c.socket.Error(err.Error())
	}

	_ = c.socket.OK()

	return join(ctx, c.socket, conn)
}

func (c *Client) handleRegister(ctx context.Context, request proto.Request) error {
	// Register the requested port
	port, err := c.net.Register(request.Port)
	if err != nil {
		return c.socket.Error(err.Error())
	}

	_ = c.socket.OK()

	// close the port when the registration connection closes
	go func() {
		defer port.Close()

		var buf [1]byte
		_, err := c.socket.Read(buf[:])
		if err == nil {
			_ = c.socket.Close()
		}
	}()

	defer c.socket.Close()
	return c.handlePort(ctx, port, request.Path)
}

func (c *Client) handlePort(ctx context.Context, port api.PortHandler, dest string) error {
	go func() {
		<-ctx.Done()
		port.Close()
	}()

	for request := range port.Requests() {
		log.Println("apps:", logfmt.ID(string(request.Caller())), "queried", request.Query())

		var rawConn net.Conn
		var err error

		// Connect to app socket
		if dest[0] == '/' {
			rawConn, err = net.Dial("unix", dest)
		} else {
			rawConn, err = net.Dial("tcp", dest)
		}
		if err != nil {
			request.Reject()
			return err
		}
		conn := proto.NewConn(rawConn)

		// Pass the request to the app
		response, err := conn.Connect(string(request.Caller()), request.Query())
		if err != nil {
			request.Reject()
			return err
		}

		// If connection was not accepted move on to the next request
		if response.Status != proto.StatusOK {
			request.Reject()
			continue
		}

		go join(ctx, request.Accept(), conn)
	}

	return nil
}
