package config

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"os"
	"path"
	"strings"
)

type Store interface {
	Read(name string) ([]byte, error)
	Write(name string, data []byte) error
	LoadYAML(name string, out interface{}) error
	StoreYAML(name string, in interface{}) error
}

type FileStore struct {
	baseDir string
}

func NewFileStore(baseDir string) (*FileStore, error) {
	return &FileStore{baseDir: baseDir}, nil
}

func (store *FileStore) Read(name string) ([]byte, error) {
	bytes, err := os.ReadFile(path.Join(store.baseDir, name))

	switch {
	case err == nil:

	case strings.Contains(err.Error(), "no such file or directory"):
		err = ErrNotFound
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

	return yaml.Unmarshal(bytes, out)
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
