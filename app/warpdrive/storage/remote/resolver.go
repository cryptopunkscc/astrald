package remote

import (
	"github.com/cryptopunkscc/astrald/app/warpdrive/api"
	"io"
)

type Resolver struct {
}

func (c Resolver) Reader(uri string) (io.ReadCloser, error) {
	//TODO implement me
	panic("implement me")
}

func (c Resolver) Info(uri string) (files []api.Info, err error) {
	//TODO implement me
	panic("implement me")
}
