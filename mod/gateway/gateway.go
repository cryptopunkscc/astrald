package gateway

import (
	"context"
	"github.com/cryptopunkscc/astrald/enc"
	"github.com/cryptopunkscc/astrald/node"
	"io"
	"sync"
)

const serviceHandle = ".gateway"
const ModuleName = "gateway"

type Gateway struct{}

func (Gateway) Run(ctx context.Context, node *node.Node) error {
	port, err := node.Ports.RegisterContext(ctx, serviceHandle)
	if err != nil {
		return err
	}

	for req := range port.Queries() {
		conn := req.Accept()

		go func() {
			nodeID, err := enc.ReadIdentity(conn)
			if err != nil {
				conn.Close()
				return
			}

			out, err := node.Query(ctx, nodeID, ".connect")
			if err != nil {
				enc.Write(conn, uint8(0))
				conn.Close()
				return
			}

			enc.Write(conn, uint8(1))

			join(ctx, conn, out)
		}()
	}

	return nil
}

func (Gateway) String() string {
	return ModuleName
}

func join(ctx context.Context, left, right io.ReadWriteCloser) error {
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		io.Copy(left, right)
		left.Close()
		wg.Done()
	}()

	go func() {
		io.Copy(right, left)
		right.Close()
		wg.Done()
	}()

	go func() {
		<-ctx.Done()
		right.Close()
	}()

	wg.Wait()
	return nil
}
