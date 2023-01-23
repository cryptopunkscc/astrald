package astral

import (
	"context"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/node"
	"io"
)

func handlerLoader(name string, handlers Handlers) node.ModuleLoader {
	return handlersModuleLoader{name: name, handlers: handlers}
}

type Handlers interface {
	Next() Handler
}

type Handler interface {
	Serve(conn Connection)
	String() string
}

type Connection interface {
	io.WriteCloser
	Read(n int) ([]byte, error)
}

type handlersModuleLoader struct {
	name     string
	handlers Handlers
}

func (loader handlersModuleLoader) Name() string {
	return loader.name
}

func (loader handlersModuleLoader) Load(node *node.Node) (node.Module, error) {
	mod := &handlersModule{node: node}
	for {
		m := loader.handlers.Next()
		if m == nil {
			break
		}
		mod.handlers = append(mod.handlers, m)
	}
	return mod, nil
}

type handlersModule struct {
	node     *node.Node
	name     string
	handlers []Handler
}

func (r handlersModule) Run(ctx context.Context) error {
	for _, handler := range r.handlers {
		handler := handler
		port, err := r.node.Ports.Register(handler.String())
		if err != nil {
			return err
		}
		go func() {
			defer port.Close()
			for q := range port.Queries() {
				conn, err := q.Accept()
				if err != nil {
					log.Error("accept query: %s", err)
					continue
				}
				finish := make(chan struct{})
				go func() {
					defer conn.Close()
					select {
					case <-ctx.Done():
					case <-finish:
					}
				}()
				go func(conn io.ReadWriteCloser) {
					defer close(finish)
					w := &ConnectionWrapper{ReadWriteCloser: conn}
					handler.Serve(w)
				}(conn)
			}
		}()
	}
	<-ctx.Done()
	return nil
}

type ConnectionWrapper struct{ io.ReadWriteCloser }

func (c *ConnectionWrapper) Read(n int) (b []byte, err error) {
	var l int
	b = make([]byte, n)
	if l, err = c.ReadWriteCloser.Read(b); err == nil {
		b = b[:l]
	}
	return
}
