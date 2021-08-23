package client

import (
	"context"
	"github.com/cryptopunkscc/astrald/api"
	"github.com/cryptopunkscc/astrald/components/fid"
	"github.com/cryptopunkscc/astrald/components/repo"
	"github.com/cryptopunkscc/astrald/components/sio"
	"github.com/cryptopunkscc/astrald/services/repo/request"
	"github.com/cryptopunkscc/astrald/services/util/connect"
	"io"
)

type client struct {
	port     string
	identity api.Identity
	ctx      context.Context
	core     api.Core
}

func (c *client) connect(request byte) (sio.ReadWriteCloser, error) {
	return connect.RemoteRequest(c.ctx, c.core, c.identity, c.port, request)
}

func (c *client) Reader(id fid.ID) (repo.Reader, error) {
	s, err := c.connect(request.Read)
	if err != nil {
		return nil, err
	}

	idBuff := id.Pack()
	_, err = s.Write(idBuff[:])
	if err != nil {
		return nil, err
	}

	return reader{s, int64(id.Size)}, nil
}

func (c *client) Writer() (w repo.Writer, err error) {
	s, err := c.connect(request.Write)
	if err != nil {
		return nil, err
	}

	return writer{s}, nil
}

func (c *client) Observer() (repo.Observer, error) {
	s, err := c.connect(request.Observe)
	if err != nil {
		return nil, err
	}

	return reader{s, -1}, nil
}

func (c *client) Map(path string) (*fid.ID, error) {
	conn, err := c.connect(request.Map)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	_, err = conn.WriteStringWithSize16(path)
	if err != nil {
		return nil, err
	}
	id, _, err := fid.Read(conn)
	if err != nil {
		return nil, err
	}
	return &id, nil
}

func (c *client) List() (reader io.ReadCloser, err error) {
	return c.connect(request.List)
}
