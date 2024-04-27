package fs

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/mod/fs"
	"github.com/cryptopunkscc/astrald/mod/storage"
	"os"
	"path/filepath"
	"sync/atomic"
)

var _ storage.Writer = &Writer{}

const tempFilePrefix = ".tmp."

type Writer struct {
	mod       *Module
	path      string
	tempID    string
	file      *os.File
	resolver  data.Resolver
	finalized atomic.Bool
}

func NewWriter(mod *Module, path string) (*Writer, error) {
	var rbytes = make([]byte, 8)
	rand.Read(rbytes)

	var tempID = tempFilePrefix + hex.EncodeToString(rbytes)

	file, err := os.Create(filepath.Join(path, tempID))
	if err != nil {
		return nil, err
	}

	resolver := data.NewResolver()

	return &Writer{
		mod:      mod,
		path:     path,
		tempID:   tempID,
		file:     file,
		resolver: resolver,
	}, nil
}

func (w *Writer) Write(p []byte) (n int, err error) {
	n, err = w.file.Write(p)

	if n > 0 {
		w.resolver.Write(p[:n])
	}

	return n, err
}

func (w *Writer) Commit() (data.ID, error) {
	if !w.finalized.CompareAndSwap(false, true) {
		return data.ID{}, errors.New("writer closed")
	}

	w.file.Close()

	dataID := w.resolver.Resolve()

	var oldPath = filepath.Join(w.path, w.tempID)
	var newPath = filepath.Join(w.path, dataID.String())

	err := os.Rename(oldPath, newPath)
	if err != nil {
		os.Remove(oldPath)
	}

	stat, err := os.Stat(newPath)
	if err != nil {
		return dataID, err
	}

	err = w.mod.db.Create(&dbLocalFile{
		Path:    newPath,
		DataID:  dataID,
		ModTime: stat.ModTime(),
	}).Error
	if err == nil {
		w.mod.events.Emit(fs.EventFileAdded{
			Path:   newPath,
			DataID: dataID,
		})
	}

	return dataID, err
}

func (w *Writer) Discard() error {
	if !w.finalized.CompareAndSwap(false, true) {
		return errors.New("writer closed")
	}

	w.file.Close()
	os.Remove(filepath.Join(w.path, w.tempID))
	return nil
}
