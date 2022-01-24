package remote

import (
	"github.com/cryptopunkscc/astrald/app/warpdrive/api"
	"io"
)

func NewResolver() api.Resolver {
	return &resolverClient{}
}

type resolverClient struct {
}

func (c *resolverClient) File(uri string) (io.ReadCloser, error) {
	//TODO implement me
	panic("implement me")
}

func (c *resolverClient) Info(uri string) (files []api.Info, err error) {
	//TODO implement me
	panic("implement me")
}
