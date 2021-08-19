package repo

import (
	"github.com/cryptopunkscc/astrald/components/fid"
	"github.com/cryptopunkscc/astrald/components/repo"
	"github.com/cryptopunkscc/astrald/components/serializer"
	"github.com/cryptopunkscc/astrald/components/storage"
	"io"
	"log"
)

const tag = "file-storage"

type adapter struct {
	storage.Storage
}

type reader struct {
	storage.FileReader
	serializer.Parser
}

type writer struct {
	fid.Resolver
	storage.FileWriter
	serializer.Formatter
}

func NewAdapter(storage storage.Storage) repo.ReadWriteRepository {
	return adapter{Storage: storage}
}

func (f adapter) Reader(id fid.ID) (repo.Reader, error) {
	r, err := f.Storage.Reader(id.String())
	if err != nil {
		return nil, err
	}

	return reader{
		FileReader: r,
		Parser:     serializer.NewReader(r),
	}, nil
}

func (f adapter) List() (reader io.ReadCloser, err error) {
	names, err := f.Storage.List()
	if err != nil {
		return
	}
	reader, pw := io.Pipe()
	go func() {
		_, err := serializer.NewWriter(pw).WriteUInt16(uint16(len(names)))
		if err != nil {
			pw.CloseWithError(err)
			return
		}
		for _, name := range names {
			parse, err := fid.Parse(name)
			if err != nil {
				_ = pw.CloseWithError(err)
				return
			}
			pack := parse.Pack()
			_, err = pw.Write(pack[:])
			if err != nil {
				_ = pw.CloseWithError(err)
				return
			}
		}
		_ = pw.Close()
	}()
	return
}

func (f adapter) Writer() (repo.Writer, error) {
	w, err := f.Storage.Writer()
	if err != nil {
		return nil, err
	}
	return &writer{
		Resolver:   fid.NewResolver(),
		FileWriter: w,
		Formatter:  serializer.NewWriter(w),
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
