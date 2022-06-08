package apphost

import (
	"context"
	"fmt"
	_node "github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/proto/apphost"
	"github.com/cryptopunkscc/astrald/storage"
	"log"
	"net"
	"os"
	"path/filepath"
	"sync"
)

type Module struct {
	node        *_node.Node
	listeners   []net.Listener
	clientConns chan net.Conn
}

const ModuleName = "apphost2"
const UnixSocketName = "apphost.sock"
const DefaultSocketDir = "/var/run/astrald"
const TCPPort = 8625
const workerCount = 32

func (mod *Module) Run(ctx context.Context, node *_node.Node) error {
	mod.node = node
	var wg sync.WaitGroup

	ports := NewPortManager()

	wg.Add(workerCount)

	for i := 0; i < workerCount; i++ {
		go func(i int) {
			defer wg.Done()
			for conn := range mod.clientConns {
				err := apphost.Serve(conn, &AppHost{
					node:  node,
					conn:  conn,
					ports: ports,
				})
				if err != nil {
					log.Printf("[apphost:%d] serve error: %s", i, err.Error())
				}
				conn.Close()
			}
		}(i)
	}

	mod.runListeners(ctx)

	wg.Wait()

	return nil
}

func (mod *Module) runListeners(ctx context.Context) {
	mod.listeners = make([]net.Listener, 0)
	mod.clientConns = make(chan net.Conn)

	if l, err := mod.listenTCP(); err != nil {
		log.Println("[apphost] tcp listen error:", err)
	} else {
		log.Println("[apphost] listen", l.Addr())
		mod.listeners = append(mod.listeners, l)
	}

	if l, err := mod.listenUnix(); err != nil {
		log.Println("[apphost] unix listen error:", err)
	} else {
		log.Println("[apphost] listen", l.Addr())
		mod.listeners = append(mod.listeners, l)
	}

	for _, listener := range mod.listeners {
		go func() {
			for {
				conn, err := listener.Accept()
				if err != nil {
					break
				}
				mod.clientConns <- conn
			}
		}()
	}

	go func() {
		<-ctx.Done()
		for _, l := range mod.listeners {
			l.Close()
		}
		close(mod.clientConns)
	}()
}

func (mod *Module) listenTCP() (net.Listener, error) {
	return net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", TCPPort))
}

func (mod *Module) listenUnix() (net.Listener, error) {
	var socketDir = DefaultSocketDir

	if fs, ok := mod.node.Store.(*storage.FilesystemStorage); ok {
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

func (Module) String() string {
	return ModuleName
}
