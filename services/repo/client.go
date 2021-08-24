package repo

import (
	"context"
	"github.com/cryptopunkscc/astrald/api"
	"github.com/cryptopunkscc/astrald/components/fid"
	"github.com/cryptopunkscc/astrald/components/repo"
	"github.com/cryptopunkscc/astrald/components/sio"
	"github.com/cryptopunkscc/astrald/services/util/connect"
	"io"
)

// =================== Constructors ====================

func NewRepoClient(
	ctx context.Context,
	core api.Core,
) repo.LocalRepository {
	return New(Port, "", ctx, core)
}

func NewFilesClient(
	ctx context.Context,
	core api.Core,
	identity api.Identity,
) repo.RemoteRepository {
	return New(FilesPort, identity, ctx, core)
}

func New(
	port string,
	identity api.Identity,
	ctx context.Context,
	core api.Core,
) repo.LocalRepository {
	return &client{
		port:     port,
		identity: identity,
		ctx:      ctx,
		core:     core,
	}
}

// =================== Client ====================

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
	s, err := c.connect(Read)
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
	s, err := c.connect(Write)
	if err != nil {
		return nil, err
	}

	return writer{s}, nil
}

func (c *client) Observer() (repo.Observer, error) {
	s, err := c.connect(Observe)
	if err != nil {
		return nil, err
	}

	return reader{s, -1}, nil
}

func (c *client) Map(path string) (*fid.ID, error) {
	conn, err := c.connect(Map)
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
	return c.connect(List)
}

// =================== Reader ====================

type reader struct {
	sio.ReadCloser
	size int64
}

func (r reader) Size() (int64, error) {
	return r.size, nil
}

// =================== Writer ====================

type writer struct {
	sio.ReadWriteCloser
}

func (w writer) Finalize() (*fid.ID, error) {
	var idBuff [fid.Size]byte
	_, err := w.Read(idBuff[:])
	if err != nil {
		return nil, err
	}
	id := fid.Unpack(idBuff)
	return &id, nil
}
