package fs

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"github.com/cryptopunkscc/astrald/mod/fs"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/object"
	"os"
	"path/filepath"
	"sync/atomic"
)

var _ objects.Writer = &Writer{}

const tempFilePrefix = ".tmp."

type Writer struct {
	mod       *Module
	path      string
	tempID    string
	file      *os.File
	resolver  object.Resolver
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

	resolver := object.NewResolver()

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

func (w *Writer) Commit() (object.ID, error) {
	if !w.finalized.CompareAndSwap(false, true) {
		return object.ID{}, errors.New("writer closed")
	}

	w.file.Close()

	objectID := w.resolver.Resolve()

	var oldPath = filepath.Join(w.path, w.tempID)
	var newPath = filepath.Join(w.path, objectID.String())

	err := os.Rename(oldPath, newPath)
	if err != nil {
		os.Remove(oldPath)
	}

	stat, err := os.Stat(newPath)
	if err != nil {
		return objectID, err
	}

	err = w.mod.db.Create(&dbLocalFile{
		Path:    newPath,
		DataID:  objectID,
		ModTime: stat.ModTime(),
	}).Error
	if err == nil {
		w.mod.events.Emit(fs.EventFileAdded{
			Path:     newPath,
			ObjectID: objectID,
		})
		w.mod.events.Emit(objects.EventObjectDiscovered{
			ObjectID: objectID,
			Zone:     net.ZoneDevice,
		})
	}

	return objectID, err
}

func (w *Writer) Discard() error {
	if !w.finalized.CompareAndSwap(false, true) {
		return errors.New("writer closed")
	}

	w.file.Close()
	os.Remove(filepath.Join(w.path, w.tempID))
	return nil
}
