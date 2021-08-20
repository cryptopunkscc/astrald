package adapter

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

type repository struct {
	port     string
	identity api.Identity
	ctx      context.Context
	core     api.Core
}

func (repo *repository) connect(request uint16) (sio.ReadWriteCloser, error) {
	return connect.LocalRequest(repo.ctx, repo.core, repo.port, request)
}

func (repo *repository) Reader(id fid.ID) (repo.Reader, error) {
	s, err := repo.connect(request.Read)
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

func (repo *repository) Writer() (w repo.Writer, err error) {
	s, err := repo.connect(request.Write)
	if err != nil {
		return nil, err
	}

	return writer{s}, nil
}

func (repo *repository) Observer() (repo.Observer, error) {
	s, err := repo.connect(request.Observe)
	if err != nil {
		return nil, err
	}

	return reader{s, -1}, nil
}

func (repo *repository) Map(path string) (*fid.ID, error) {
	conn, err := repo.connect(request.Map)
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

func (repo *repository) List() (reader io.ReadCloser, err error) {
	return repo.connect(request.List)
}
