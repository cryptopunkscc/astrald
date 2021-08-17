package repo

import (
	"github.com/cryptopunkscc/astrald/components/fid"
	"github.com/cryptopunkscc/astrald/components/repo"
	"github.com/cryptopunkscc/astrald/components/storage"
	"log"
)

const tag = "file-storage"

type adapter struct {
	delegate storage.Storage
}

type reader struct {
	delegate storage.FileReader
}

type writer struct {
	resolver fid.Resolver
	delegate storage.FileWriter
}

func NewAdapter(storage storage.Storage) repo.Repository {
	return adapter{delegate: storage}
}

func (f adapter) Reader(id fid.ID) (repo.Reader, error) {
	r, err := f.delegate.Reader(id.String())
	if err != nil {
		return nil, err
	}
	return &reader{delegate: r}, nil
}

func (r reader) Size() (int64, error) {
	return r.delegate.Size()
}

func (f adapter) Writer() (repo.Writer, error) {
	w, err := f.delegate.Writer()
	if err != nil {
		return nil, err
	}
	return &writer{
		resolver: fid.NewResolver(),
		delegate: w,
	}, nil
}

func (r reader) Read(p []byte) (int, error) {
	return r.delegate.Read(p)
}

func (r reader) Close() error {
	return r.delegate.Close()
}

func (w writer) Write(p []byte) (int, error) {
	_, _ = w.resolver.Write(p)
	return w.delegate.Write(p)
}

func (w writer) Finalize() (*fid.ID, error) {
	err := w.delegate.Sync()
	if err != nil {
		log.Println(err)
		return nil, err
	}
	id := w.resolver.Resolve()
	err = w.delegate.Rename(id.String())
	if err != nil {
		log.Println(tag, "Cannot rename file", err)
		return nil, err
	}
	err = w.delegate.Close()
	if err != nil {
		log.Println(tag, "Cannot close file", err)
		return nil, err
	}
	return &id, nil
}
