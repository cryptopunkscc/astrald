package repo

import (
	"context"
	"github.com/cryptopunkscc/astrald/api"
	"github.com/cryptopunkscc/astrald/components/fid"
	"github.com/cryptopunkscc/astrald/components/repo"
	"github.com/cryptopunkscc/astrald/components/serialize"
	"io"
)

type repository struct {
	identity api.Identity
	ctx      context.Context
	core     api.Core
}

type reader struct {
	io.ReadCloser
	size int64
}

type writer struct {
	io.ReadWriteCloser
}

func NewProxy(
	identity api.Identity,
	ctx context.Context,
	core api.Core,
) repo.ReadWriteRepository {
	return repository{
		identity: identity,
		ctx:      ctx,
		core:     core,
	}
}

func (repo repository) connect(request byte) (*serialize.Serializer, error) {
	stream, err := repo.core.Network().Connect(repo.identity, Port)
	if err != nil {
		return nil, nil
	}
	go func() {
		<-repo.ctx.Done()
		_ = stream.Close()
	}()
	s := serialize.NewSerializer(stream)
	err = s.WriteByte(request)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (repo repository) Reader(id fid.ID) (repo.Reader, error) {
	s, err := repo.connect(RequestRead)
	if err != nil {
		return nil, err
	}

	return reader{s, int64(id.Size)}, nil
}

func (repo repository) Writer() (w repo.Writer, err error) {
	s, err := repo.connect(RequestWrite)
	if err != nil {
		return nil, err
	}

	return writer{s}, nil
}

func (repo repository) Observer() (repo.Observer, error) {
	s, err := repo.connect(RequestObserve)
	if err != nil {
		return nil, err
	}

	return reader{s, -1}, nil
}

func (repo repository) List() (reader io.ReadCloser, err error) {
	return repo.connect(RequestList)
}

func (r reader) Size() (int64, error) {
	return r.size, nil
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
