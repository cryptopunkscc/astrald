package warpdrive

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/lib/wrapper"
	"io"
	"log"
)

type Client struct {
	wrapper.Api
	*log.Logger
	*cslq.Endec
	Conn       io.ReadWriteCloser
	LocalNode  string
	RemoteNode string
}

type RemoteClient struct{ Client }
type LocalClient struct{ Client }

func NewClient(api wrapper.Api) Client {
	c := Client{Logger: log.Default(), Api: api}
	localNode, err := c.Resolve("localnode")
	if err != nil {
		log.Panicln("Cannot resolve local node id", err)
	}
	c.LocalNode = localNode.String()
	return c
}

// ConnectLocal connects to local warpdrive service
func (c Client) ConnectLocal() (lc LocalClient, err error) {
	if err = c.connect(id.Identity{}); err != nil {
		return
	}
	lc = LocalClient{c}
	return
}

// ConnectRemote connects to remote warpdrive service
func (c Client) ConnectRemote(identity id.Identity) (rc RemoteClient, err error) {
	if err = c.connect(identity); err != nil {
		return
	}
	rc = RemoteClient{c}
	return
}

func (c *Client) connect(identity id.Identity) (err error) {
	c.Logger = NewLogger("[CLIENT]", Port)
	// Connect to local service
	conn, err := c.Query(identity, Port)
	if err != nil {
		c.Println("Cannot connect to service", err)
		return
	}
	c.Conn = conn
	c.Endec = cslq.NewEndec(conn)
	if identity.IsZero() {
		c.RemoteNode = c.LocalNode
	} else {
		c.RemoteNode = identity.String()
	}
	return
}

func (c *Client) Close() (err error) {
	err = c.Encode("c", cmdClose)
	_ = c.Conn.Close()
	return
}
