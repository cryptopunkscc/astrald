package astral

import (
	"context"
	"github.com/cryptopunkscc/astrald/node"
	"io"
	"log"
)

func handlerRunner(name string, handlers Handlers) (moduleRunner node.ModuleRunner) {
	r := &handlersModuleRunner{}
	r.name = name
	moduleRunner = r
	for {
		m := handlers.Next()
		if m == nil {
			return
		}
		r.handlers = append(r.handlers, m)
	}
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

type handlersModuleRunner struct {
	name     string
	handlers []Handler
}

func (r handlersModuleRunner) String() string {
	return r.name
}

func (r handlersModuleRunner) Run(ctx context.Context, n *node.Node) error {
	for _, handler := range r.handlers {
		handler := handler
		port, err := n.Ports.Register(handler.String())
		if err != nil {
			return err
		}
		go func() {
			defer port.Close()
			for q := range port.Queries() {
				conn, err := q.Accept()
				if err != nil {
					log.Println("Cannot accept query", err)
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
