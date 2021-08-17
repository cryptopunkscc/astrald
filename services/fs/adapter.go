package fs

import (
	"github.com/cryptopunkscc/astrald/components/fid"
	"log"
)

const tag = "file-storage"

type adapter struct {
	delegate Storage
}

type reader struct {
	delegate FileReader
}

type writer struct {
	resolver fid.Resolver
	delegate FileWriter
}

func NewAdapter(storage Storage) Repository {
	return adapter{delegate: storage}
}

func (f adapter) Reader(id fid.ID) (Reader, error) {
	r, err := f.delegate.Reader(id.String())
	if err != nil {
		return nil, err
	}
	return &reader{delegate: r}, nil
}

func (r reader) Size() (int64, error) {
	return r.delegate.Size()
}

func (f adapter) Writer() (Writer, error) {
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
