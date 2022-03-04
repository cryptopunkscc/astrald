package apphost

import (
	"context"
	"fmt"
	_node "github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/storage"
	"log"
	"net"
	"os"
	"path/filepath"
)

type AppHost struct {
	node        *_node.Node
	listeners   []net.Listener
	clientConns chan net.Conn
}

const ModuleName = "apphost"
const UnixSocketName = "apphost.sock"
const DefaultSocketDir = "/var/run/astrald"
const TCPPort = 8625

func (host *AppHost) Run(ctx context.Context, node *_node.Node) error {
	host.node = node

	host.runListeners(ctx)

	for conn := range host.clientConns {
		ServeClient(ctx, conn, node)
	}

	return nil
}

func (host *AppHost) runListeners(ctx context.Context) {
	host.listeners = make([]net.Listener, 0)
	host.clientConns = make(chan net.Conn)

	if l, err := host.listenTCP(); err != nil {
		log.Println("[apphost] tcp listen error:", err)
	} else {
		log.Println("[apphost] listen", l.Addr())
		host.listeners = append(host.listeners, l)
	}

	if l, err := host.listenUnix(); err != nil {
		log.Println("[apphost] unix listen error:", err)
	} else {
		log.Println("[apphost] listen", l.Addr())
		host.listeners = append(host.listeners, l)
	}

	for _, listener := range host.listeners {
		go func() {
			for {
				conn, err := listener.Accept()
				if err != nil {
					break
				}
				host.clientConns <- conn
			}
		}()
	}

	go func() {
		<-ctx.Done()
		for _, l := range host.listeners {
			l.Close()
		}
		close(host.clientConns)
	}()
}

func (host *AppHost) listenTCP() (net.Listener, error) {
	return net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", TCPPort))
}

func (host *AppHost) listenUnix() (net.Listener, error) {
	var socketDir = DefaultSocketDir

	if fs, ok := host.node.Store.(*storage.FilesystemStorage); ok {
		socketDir = fs.Root()
	}

	socketPath := filepath.Join(socketDir, UnixSocketName)

	if info, err := os.Stat(socketPath); err == nil {
		if !info.Mode().IsRegular() {
			log.Println("[apphost] stale unix socket found, removing...")
			err := os.Remove(socketPath)
			if err != nil {
				return nil, err
			}
		}
	}

	return net.Listen("unix", socketPath)
}

func (AppHost) String() string {
	return ModuleName
}
