package apphost

import (
	"context"
	"fmt"
	_log "github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/proto/apphost"
	"net"
	"os"
	"path/filepath"
	"sync"
)

type Module struct {
	node        *node.Node
	listeners   []net.Listener
	clientConns chan net.Conn
}

var log = _log.Tag(ModuleName)

const UnixSocketName = "apphost.sock"
const DefaultSocketDir = "/var/run/astrald"
const TCPPort = 8625
const workerCount = 32

func (mod *Module) Run(ctx context.Context) error {
	var wg sync.WaitGroup

	ports := NewPortManager()
	conns := mod.accept(ctx)

	wg.Add(workerCount)
	for i := 0; i < workerCount; i++ {
		go func(i int) {
			defer wg.Done()
			for conn := range conns {
				cctx, cancel := context.WithCancel(ctx)

				go func() {
					select {
					case <-ctx.Done():
						conn.Close()
					case <-cctx.Done():
					}
				}()

				err := apphost.Serve(conn, &AppHost{
					node:  mod.node,
					conn:  conn,
					ports: ports,
				})
				if err != nil {
					log.Error("worker %d serve error: %s", i, err.Error())
				}
				conn.Close()
				cancel()
			}
		}(i)
	}

	wg.Wait()

	return nil
}

func (mod *Module) accept(ctx context.Context) <-chan net.Conn {
	if l, err := mod.listenTCP(); err != nil {
		log.Error("tcp listen error: %s", err)
	} else {
		log.Log("listen %s", log.Em(l.Addr().String()))
		mod.listeners = append(mod.listeners, l)
	}

	if l, err := mod.listenUnix(); err != nil {
		log.Error("unix listen error:", err)
	} else {
		log.Log("listen %s", log.Em(l.Addr().String()))
		mod.listeners = append(mod.listeners, l)
	}

	var ch = make(chan net.Conn)
	var wg sync.WaitGroup

	for _, listener := range mod.listeners {
		listener := listener
		wg.Add(1)

		go func() {
			defer wg.Done()

			for {
				conn, err := listener.Accept()
				if err != nil {
					break
				}
				ch <- conn
			}
		}()
		go func() {
			<-ctx.Done()
			listener.Close()
		}()
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	return ch
}

func (mod *Module) listenTCP() (net.Listener, error) {
	return net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", TCPPort))
}

func (mod *Module) listenUnix() (net.Listener, error) {

	var socketDir = mod.node.RootDir()
	if socketDir == "" {
		socketDir = DefaultSocketDir
	}

	socketPath := filepath.Join(socketDir, UnixSocketName)

	if info, err := os.Stat(socketPath); err == nil {
		if !info.Mode().IsRegular() {
			log.Log("stale unix socket found, removing...")
			err := os.Remove(socketPath)
			if err != nil {
				return nil, err
			}
		}
	}

	return net.Listen("unix", socketPath)
}
