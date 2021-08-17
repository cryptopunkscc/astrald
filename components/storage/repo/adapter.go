package repo

import (
	"github.com/cryptopunkscc/astrald/components/fid"
	"github.com/cryptopunkscc/astrald/components/repo"
	"github.com/cryptopunkscc/astrald/components/storage"
	"log"
)

const tag = "file-storage"

type adapter struct {
	storage.Storage
}

type reader struct {
	storage.FileReader
}

type writer struct {
	fid.Resolver
	storage.FileWriter
}

func NewAdapter(storage storage.Storage) repo.Repository {
	return adapter{Storage: storage}
}

func (f adapter) Reader(id fid.ID) (repo.Reader, error) {
	r, err := f.Storage.Reader(id.String())
	if err != nil {
		return nil, err
	}
	return &reader{FileReader: r}, nil
}

func (f adapter) Writer() (repo.Writer, error) {
	w, err := f.Storage.Writer()
	if err != nil {
		return nil, err
	}
	return &writer{
		Resolver: fid.NewResolver(),
		FileWriter: w,
	}, nil
}

func (w writer) Write(p []byte) (int, error) {
	_, _ = w.Resolver.Write(p)
	return w.FileWriter.Write(p)
}

func (w writer) Finalize() (*fid.ID, error) {
	err := w.Sync()
	if err != nil {
		log.Println(err)
		return nil, err
	}
	id := w.Resolve()
	err = w.Rename(id.String())
	if err != nil {
		log.Println(tag, "Cannot rename file", err)
		return nil, err
	}
	err = w.Close()
	if err != nil {
		log.Println(tag, "Cannot close file", err)
		return nil, err
	}
	return &id, nil
}
