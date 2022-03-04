package apphost

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/mod/apphost/proto"
	_node "github.com/cryptopunkscc/astrald/node"
	"io"
)

type Server struct {
	*cslq.Endec
	conn   io.ReadWriteCloser
	node   *_node.Node
	cancel context.CancelFunc
}

func ServeClient(ctx context.Context, conn io.ReadWriteCloser, node *_node.Node) (*Server, error) {
	ctx, cancel := context.WithCancel(ctx)

	c := &Server{
		conn:   conn,
		node:   node,
		Endec:  cslq.NewEndec(conn),
		cancel: cancel,
	}

	go c.handle(ctx)

	return c, nil
}

func (c *Server) handle(ctx context.Context) error {
	var reqType int

	if err := c.Decode("c", &reqType); err != nil {
		return err
	}

	switch reqType {
	case proto.RequestInfo:
		return c.handleInfo(ctx)

	case proto.RequestDialKey:
		return c.handleDialKey(ctx)

	case proto.RequestDialString:
		return c.handleDialString(ctx)

	case proto.RequestRegister:
		return c.handleRegister(ctx)

	case proto.RequestGetNodeName:
		return c.handleGetNodeName(ctx)

	default:
		c.conn.Close() // ignore unknown requests
		return errors.New("invalid request")
	}
}

func (c *Server) closeWithError(err error) error {
	c.Encode("c", proto.ResponseRejected)
	c.conn.Close()
	return err
}
