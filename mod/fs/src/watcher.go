package fs

import (
	"errors"
	"github.com/cryptopunkscc/astrald/sig"
	"github.com/fsnotify/fsnotify"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"
)

const DefaultWriteTimeout = 3 * time.Second

type Watcher struct {
	OnWrite       func(string)
	onWriteDone   func(string)
	OnFileCreated func(string)
	OnDirCreated  func(string)
	OnRenamed     func(string)
	OnRemoved     func(string)
	OnChmod       func(string)
	OnError       func(error)
	WriteTimeout  time.Duration

	watcher  *fsnotify.Watcher
	timeouts map[string]*time.Time
	mu       sync.Mutex
}

func NewWatcher() (*Watcher, error) {
	var err error
	var w = &Watcher{
		WriteTimeout: DefaultWriteTimeout,
		timeouts:     map[string]*time.Time{},
	}

	w.watcher, err = fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	go w.worker()

	return w, nil
}

func (w *Watcher) Add(path string, tree bool) error {
	if slices.Contains(w.watcher.WatchList(), path) {
		return errors.New("already added")
	}

	stat, err := os.Stat(path)
	if err != nil {
		return err
	}
	if !stat.IsDir() {
		return errors.New("not a directory")
	}

	err = w.watcher.Add(path)
	if err != nil {
		return err
	}

	if !tree {
		return nil
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		w.Add(filepath.Join(path, entry.Name()), tree)
	}

	return nil
}

func (w *Watcher) Remove(path string, tree bool) error {
	err := w.watcher.Remove(path)
	if err != nil {
		return err
	}

	if !tree {
		return nil
	}

	for _, p := range w.watcher.WatchList() {
		if strings.HasPrefix(p, path+"/") {
			w.Remove(p, true)
		}
	}

	return nil
}

func (w *Watcher) onWrite(path string) {
	w.mu.Lock()
	defer w.mu.Unlock()

	t, found := w.timeouts[path]
	if found {
		*t = time.Now().Add(w.WriteTimeout)
		return
	}

	w.timeouts[path] = &time.Time{}
	*w.timeouts[path] = time.Now().Add(w.WriteTimeout)

	sig.At(w.timeouts[path], &w.mu, func() {
		w.mu.Lock()
		delete(w.timeouts, path)
		w.mu.Unlock()

		if w.onWriteDone != nil {
			w.onWriteDone(path)
		}
	})
}

func (w *Watcher) onRemoved(path string) {
	w.mu.Lock()
	defer w.mu.Unlock()

	t, found := w.timeouts[path]
	if !found {
		return
	}

	*t = time.Time{}
}

func (w *Watcher) onRenamed(path string) {
	w.onRemoved(path)
}

func (w *Watcher) worker() {
	for {
		select {
		case event, ok := <-w.watcher.Events:
			if !ok {
				return
			}

			switch {
			case event.Op.Has(fsnotify.Write):
				w.onWrite(event.Name)
				if w.OnWrite != nil {
					w.OnWrite(event.Name)
				}

			case event.Op.Has(fsnotify.Rename):
				w.onRenamed(event.Name)
				if w.OnRenamed != nil {
					w.OnRenamed(event.Name)
				}

			case event.Op.Has(fsnotify.Create):
				info, err := os.Stat(event.Name)
				switch {
				case err != nil:
				case info.Mode().IsRegular():
					if w.OnFileCreated != nil {
						w.onWrite(event.Name)
						w.OnFileCreated(event.Name)
					}

				case info.Mode().IsDir():
					if w.OnDirCreated != nil {
						w.OnDirCreated(event.Name)
					}
				}

			case event.Op.Has(fsnotify.Chmod):
				if w.OnChmod != nil {
					w.OnChmod(event.Name)
				}

			case event.Op.Has(fsnotify.Remove):
				w.onRemoved(event.Name)
				if w.OnRemoved != nil {
					w.OnRemoved(event.Name)
				}
			}

		case e, ok := <-w.watcher.Errors:
			if !ok {
				return
			}

			if w.OnError != nil {
				w.OnError(e)
			}
		}
	}
}
