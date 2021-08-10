package tcp

import (
	"context"
	"fmt"
	"github.com/cryptopunkscc/astrald/api"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/services/appsupport/proto"
	"io"
	"log"
	"net"
	"os"
)

const tcpAppSupportAddress = "127.0.0.1:8625"

func init() {
	_ = node.RegisterService("tcp", &Runner{})
}

type Runner struct{}

func (runner *Runner) Run(ctx context.Context, core api.Core) error {
	network := core.Network()

	go func() {
		p, _ := network.Register("apps-tcp")

		for r := range p.Requests() {
			c := r.Accept()
			var buf [4096]byte
			c.Read(buf[:])
		}
	}()

	// Prepare the control socket
	ctl, err := makeSocket(tcpAppSupportAddress)
	if err != nil {
		return fmt.Errorf("error creating ctl socket: %v", err)
	}

	log.Println("appsupport socket:", ctl.Addr().String())

	defer func() {
		ctl.Close()
		os.Remove(ctl.Addr().String())
	}()

	go func() {
		<-ctx.Done()
		ctl.Close()
	}()

	for inConn := range accept(ctl) {
		go handleCtl(network, inConn)
	}

	return nil
}

func handleCtl(network api.Network, client *proto.Socket) error {
	request, err := client.ReadRequest()
	if err != nil {
		return err
	}

	switch request.Type {
	case proto.RequestConnect:
		conn, err := network.Connect(api.Identity(request.Identity), request.Port)
		if err != nil {
			return client.Error(err.Error())
		}
		client.OK()
		join(client, conn)
	case proto.RequestRegister:
		// Register the requested port
		port, err := network.Register(request.Port)
		if err != nil {
			return client.Error(err.Error())
		}
		client.OK()
		go handlePort(port, request.Path)
		go func() {
			for {
				var buf [4096]byte
				_, err := client.Read(buf[:])
				if err != nil {
					port.Close()
					return
				}
			}
		}()
	default:
		client.Close() // ignore unknown requests
	}
	return nil
}

func handlePort(port api.PortHandler, path string) error {
	defer port.Close()

	for request := range port.Requests() {
		log.Println(request.Caller(), "calling", request.Query())

		socket, err := net.Dial("tcp", path)
		if err != nil {
			request.Reject()
			return err
		}

		outConn := proto.NewJsonSocket(socket)
		res, err := outConn.Connect(string(request.Caller()), request.Query())
		if err != nil {
			request.Reject()
			return err
		}

		// If connection was not accepted move on to the next request
		if res.Status != proto.StatusOK {
			request.Reject()
			outConn.Close()
			continue
		}

		inConn := request.Accept()
		join(inConn, outConn)
	}
	return nil
}

// accept takes new connections from a listener and returns them through a channel
func accept(l net.Listener) <-chan *proto.Socket {
	output := make(chan *proto.Socket)

	go func() {
		defer close(output)
		for {
			conn, err := l.Accept()
			if err != nil {
				return
			}

			output <- proto.NewJsonSocket(conn)
		}
	}()

	return output
}

func join(a, b io.ReadWriteCloser) {
	go func() {
		io.Copy(a, b)
		a.Close()
	}()
	go func() {
		io.Copy(b, a)
		b.Close()
	}()
}

// makeSocket creates a new unix socket and returns its listener. If no name is provided a random one is used.
func makeSocket(address string) (net.Listener, error) {
	listen, err := net.Listen("tcp", address)
	if err != nil {
		return nil, err
	}

	return listen, nil
}
