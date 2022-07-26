package astral

import (
	"context"
	"github.com/cryptopunkscc/astrald/node"
	"io"
	"log"
)

func androidRunners(mods Modules) (r []node.ModuleRunner) {
	for {
		m := mods.Next()
		if m == nil {
			return
		}
		r = append(r, androidModuleRunner{m})
	}
}

type Modules interface {
	Next() Module
}

type Module interface {
	Serve(conn Connection)
	String() string
}

type Connection interface {
	io.WriteCloser
	Read(n int) ([]byte, error)
}

type androidModuleRunner struct{ Module }

func (a androidModuleRunner) Run(ctx context.Context, n *node.Node) error {

	port, err := n.Ports.Register(a.String())
	if err != nil {
		return err
	}
	defer port.Close()

	go func() {
		for q := range port.Queries() {
			conn, err := q.Accept()
			if err != nil {
				log.Println("Cannot accept query", err)
				continue
			}
			go func(conn io.ReadWriteCloser) {
				defer conn.Close()
				w := &ConnectionWrapper{ReadWriteCloser: conn}
				a.Serve(w)
			}(conn)
		}
	}()

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
