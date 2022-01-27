package handle

import (
	"github.com/cryptopunkscc/astrald/app/warpdrive/api"
	astral "github.com/cryptopunkscc/astrald/mod/apphost/client"
	"io"
	"log"
)

var _ api.Sender = &Client{}
var _ api.Recipient = &Client{}

type Client struct {
	Sender
	Recipient
}

type Sender struct{ *astralApi }
type Recipient struct{ *astralApi }

func NewClient(api astral.Api) Client {
	core := &astralApi{log.Default(), api}
	return Client{Sender{core}, Recipient{core}}
}

type astralApi struct {
	*log.Logger
	astral.Api
}

func (client *astralApi) query(port string) (conn io.ReadWriteCloser, err error) {
	client.Logger = api.NewLogger("<", port)
	// Connect to local service
	conn, err = client.Query("", port)
	if err != nil {
		client.Println("Cannot connect to service", err)
	}
	return
}
