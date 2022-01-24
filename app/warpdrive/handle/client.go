package handle

import (
	"github.com/cryptopunkscc/astrald/app/warpdrive/api"
	"github.com/cryptopunkscc/astrald/app/warpdrive/service"
	astral "github.com/cryptopunkscc/astrald/mod/apphost/client"
	"io"
	"log"
)

func NewClient(astralApi astral.Api) api.Client {
	c := client{log.Default(), astralApi}
	return &apiClient{
		sender{c},
		recipient{c},
	}
}

type apiClient struct {
	sender
	recipient
}
type sender struct {
	client
}

type recipient struct {
	client
}

type client struct {
	*log.Logger
	astral.Api
}

func (s *apiClient) Sender() api.Sender {
	return &s.sender
}

func (s *apiClient) Recipient() api.Recipient {
	return &s.recipient
}

func (client *client) query(port string) (conn io.ReadWriteCloser, err error) {
	client.Logger = service.NewLogger("<", port)
	// Connect to local service
	conn, err = client.Query("", port)
	if err != nil {
		client.Println("Cannot connect to service", err)
	}
	return
}
