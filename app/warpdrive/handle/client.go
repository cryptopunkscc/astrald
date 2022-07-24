package handle

import (
	"github.com/cryptopunkscc/astrald/app/warpdrive/api"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/lib/astral"
	"io"
	"log"
)

var _ api.Client = Client{}

type Client struct {
	*log.Logger
	astral.Api
	localNode string
}

func NewClient(api astral.Api) Client {
	c := Client{Logger: log.Default(), Api: api}
	localNode, err := c.Resolve("localnode")
	if err != nil {
		log.Panicln("Cannot resolve local node id", err)
	}
	c.localNode = localNode.String()
	return c
}

func (c *Client) query(port string) (conn io.ReadWriteCloser, err error) {
	c.Logger = api.NewLogger("<", port)
	// Connect to local service
	conn, err = c.Query(id.Identity{}, port)
	if err != nil {
		c.Println("Cannot connect to service", err)
		return nil, err
	}
	return
}
