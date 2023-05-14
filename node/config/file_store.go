package config

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"os"
	"path"
	"strings"
)

type FileStore struct {
	baseDir string
	Errorv  func(level int, fmt string, v ...any)
}

func NewFileStore(baseDir string) (*FileStore, error) {
	return &FileStore{
		baseDir: baseDir,
		Errorv: func(_ int, f string, v ...any) {
			if !strings.HasSuffix(f, "\n") {
				f = f + "\n"
			}
			fmt.Printf(f, v...)
		},
	}, nil
}

func (store *FileStore) Read(name string) ([]byte, error) {
	bytes, err := os.ReadFile(path.Join(store.baseDir, name))

	switch {
	case err == nil:

	case strings.Contains(err.Error(), "no such file or directory"):
		store.Errorv(2, "config %s not found", name)
		err = ErrNotFound

	default:
		store.Errorv(1, "error reading config %s: %s", name, err)
	}

	return bytes, err
}

func (store *FileStore) Write(name string, data []byte) error {
	if s, _ := os.Stat(store.baseDir); s == nil || !s.IsDir() {
		if err := os.MkdirAll(store.baseDir, 0700); err != nil {
			return fmt.Errorf("cannot create config directory: %w", err)
		}
	}

	return os.WriteFile(path.Join(store.baseDir, name), data, 0600)
}

func (store *FileStore) LoadYAML(name string, out interface{}) error {
	if !strings.HasSuffix(strings.ToLower(name), ".yaml") {
		name = name + ".yaml"
	}

	bytes, err := store.Read(name)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(bytes, out)
	if err != nil {
		store.Errorv(1, "error parsing config %s: %s", name, err)
	}
	return err
}

func (store *FileStore) StoreYAML(name string, in interface{}) error {
	if !strings.HasSuffix(strings.ToLower(name), ".yaml") {
		name = name + ".yaml"
	}

	bytes, err := yaml.Marshal(in)
	if err != nil {
		return err
	}

	return store.Write(name, bytes)
}
