package handle

import (
	"github.com/cryptopunkscc/astrald/app/warpdrive/api"
	"github.com/cryptopunkscc/astrald/lib/astral"
	"io"
	"log"
)

var _ api.Client = Client{}

type Client struct {
	*log.Logger
	astral.Api
}

func NewClient(api astral.Api) Client {
	return Client{log.Default(), api}
}

func (c *Client) query(port string) (conn io.ReadWriteCloser, err error) {
	c.Logger = api.NewLogger("<", port)
	// Connect to local service
	conn, err = c.Query("", port)
	if err != nil {
		c.Println("Cannot connect to service", err)
	}
	return
}
