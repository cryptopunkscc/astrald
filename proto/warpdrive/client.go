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

func NewClient(api wrapper.Api) Client {
	return Client{Logger: log.Default(), Api: api}
}

// Connect to warpdrive service
func (c Client) Connect(identity id.Identity, port string) (client Client, err error) {
	c.Logger = NewLogger("[CLIENT]", port)
	// Resolve local id
	localId, err := c.Resolve("localnode")
	if err != nil {
		c.Println("Cannot resolve local node id", err)
		return
	}
	c.LocalNode = localId.String()
	// Connect to local service
	conn, err := c.Query(identity, port)
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
	client = c
	return
}

func (c Client) Close() (err error) {
	err = c.Encode("c", cmdClose)
	_ = c.Conn.Close()
	return
}
