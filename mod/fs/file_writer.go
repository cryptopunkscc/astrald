package fs

import (
	"crypto/rand"
	"encoding/hex"
	"github.com/cryptopunkscc/astrald/data"
	storage "github.com/cryptopunkscc/astrald/mod/storage/api"
	"os"
	"path/filepath"
)

type FileWriter struct {
	path     string
	tempID   string
	file     *os.File
	resolver data.Resolver
	store    *StorerService
}

func NewFileWriter(parent *StorerService, path string) (*FileWriter, error) {
	var rbytes = make([]byte, 8)
	rand.Read(rbytes)

	var tempID = ".tmp." + hex.EncodeToString(rbytes)

	file, err := os.Create(filepath.Join(path, tempID))
	if err != nil {
		return nil, err
	}

	resolver := data.NewResolver()

	return &FileWriter{
		path:     path,
		tempID:   tempID,
		file:     file,
		resolver: resolver,
		store:    parent,
	}, nil
}

func (w *FileWriter) Write(p []byte) (n int, err error) {
	n, err = w.file.Write(p)

	if n > 0 {
		w.resolver.Write(p[:n])
	}

	return n, err
}

func (w *FileWriter) Commit() (data.ID, error) {
	w.file.Close()

	dataID := w.resolver.Resolve()

	var oldPath = filepath.Join(w.path, w.tempID)
	var newPath = filepath.Join(w.path, dataID.String())

	err := os.Rename(oldPath, newPath)
	if err != nil {
		os.Remove(oldPath)
	}

	info, err := os.Stat(newPath)
	if err != nil {
		return data.ID{}, err
	}

	if w.store != nil {
		w.store.events.Emit(storage.EventDataAdded{
			ID:        dataID,
			IndexedAt: info.ModTime(),
		})
	}

	return dataID, err
}

func (w *FileWriter) Discard() {
	w.file.Close()
	os.Remove(filepath.Join(w.path, w.tempID))
}
