package appsupport

import (
	"context"
	"fmt"
	"github.com/cryptopunkscc/astrald/api"
	"github.com/cryptopunkscc/astrald/logfmt"
	"github.com/cryptopunkscc/astrald/services/appsupport/proto"
	"github.com/google/uuid"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
)

// Generic, system-wide directory used as a fallback when no other setting exists
const defaultAstralDir = "/var/run/astrald"

// Name of the control socket
const ctlSocket = "ctl.sock"

type AppSupport struct {
	network api.Network
}

func (apps *AppSupport) Run(ctx context.Context) error {
	// Prepare the control socket
	ctl, err := makeSocket(ctlSocket)
	if err != nil {
		return fmt.Errorf("error creating ctl socket: %v", err)
	}

	log.Println("apps socket:", ctl.Addr().String())

	defer func() {
		ctl.Close()
		os.Remove(ctl.Addr().String())
	}()

	go func() {
		<-ctx.Done()
		ctl.Close()
	}()

	for inConn := range accept(ctl) {
		go apps.handleCtl(inConn)
	}

	return nil
}

func (apps *AppSupport) handleCtl(client *proto.Socket) error {
	request, err := client.ReadRequest()
	if err != nil {
		return err
	}

	switch request.Type {
	case proto.RequestConnect:
		conn, err := apps.network.Connect(api.Identity(request.Identity), request.Port)
		if err != nil {
			return client.Error(err.Error())
		}
		client.OK()
		join(client, conn)
	case proto.RequestRegister:
		// Register the requested port
		port, err := apps.network.Register(request.Port)
		if err != nil {
			return client.Error(err.Error())
		}
		client.OK()
		go apps.handlePort(port, request.Path)
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

func (apps *AppSupport) handlePort(port api.PortHandler, path string) error {
	defer port.Close()

	for request := range port.Requests() {
		log.Println(logfmt.ID(string(request.Caller())), "calling", request.Query())

		unix, err := net.Dial("unix", path)
		if err != nil {
			request.Reject()
			return err
		}

		outConn := proto.NewSocket(unix)
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

			output <- proto.NewSocket(conn)
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
func makeSocket(name string) (net.Listener, error) {
	if name == "" {
		name = uuid.New().String()
	}

	fullPath := socketPath(name)

	listen, err := net.Listen("unix", fullPath)
	if err != nil {
		return nil, err
	}

	return listen, nil
}

func socketPath(name string) string {
	return filepath.Join(astralDir(), name)
}

// TODO: This should be injected by the node
func astralDir() string {
	cfgDir, err := os.UserConfigDir()
	if err != nil {
		return defaultAstralDir
	}

	dir := filepath.Join(cfgDir, "astrald")
	err = os.MkdirAll(dir, 0700)
	if err != nil {
		fmt.Println("astrald dir erreror:", err)
		return defaultAstralDir
	}

	return dir
}
