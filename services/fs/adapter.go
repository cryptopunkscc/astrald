package fs

import "log"

type adapter struct {
	storage Storage
}

type reader struct {
	delegate FileReader
}

type writer struct {
	resolver Resolver
	delegate FileWriter
}

func StorageAdapter(storage Storage) Repository {
	return adapter{storage: storage}
}

func (f adapter) Reader(id ID) (Reader, error) {
	r, err := f.storage.Reader(id.String())
	if err != nil {
		return nil, err
	}
	return &reader{delegate: r}, nil
}

func (f adapter) Writer() (Writer, error) {
	w, err := f.storage.Writer()
	if err != nil {
		return nil, err
	}
	return &writer{
		resolver: NewResolver(),
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

func (w writer) Finalize() (*ID, error) {
	err := w.delegate.Sync()
	if err != nil {
		log.Println(err)
		return nil, err
	}
	id := w.resolver.Resolve()
	log.Println("Resolved file id", id.String())
	err = w.delegate.Rename(id.String())
	if err != nil {
		log.Println("Cannot rename file", err)
		return nil, err
	}
	err = w.delegate.Close()
	if err != nil {
		log.Println("Cannot close file", err)
		return nil, err
	}
	return &id, nil
}
