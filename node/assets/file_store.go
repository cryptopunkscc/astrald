package assets

import (
	"errors"
	"fmt"
	"github.com/cryptopunkscc/astrald/log"
	"gopkg.in/yaml.v2"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"os"
	"path"
	"path/filepath"
	"strings"
)

var _ Store = &FileStore{}

type FileStore struct {
	baseDir  string
	log      *log.Logger
	keyStore KeyStore
}

func NewFileStore(baseDir string, log *log.Logger) (*FileStore, error) {
	return &FileStore{
		baseDir: baseDir,
		log:     log,
	}, nil
}

func (store *FileStore) Read(name string) ([]byte, error) {
	bytes, err := os.ReadFile(path.Join(store.baseDir, name))

	switch {
	case err == nil:

	case strings.Contains(err.Error(), "no such file or directory"):
		err = ErrNotFound

	default:
		store.log.Errorv(1, "error reading config %s: %s", name, err)
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
		store.log.Errorv(1, "error parsing config %s: %s", name, err)
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

var SqliteOpen func(string) gorm.Dialector

func (store *FileStore) OpenDB(name string) (*gorm.DB, error) {
	if name == "" {
		return nil, errors.New("invalid name")
	}
	if !strings.HasSuffix(name, ".db") {
		name = name + ".db"
	}
	return gorm.Open(
		SqliteOpen(filepath.Join(store.baseDir, name)),
		&gorm.Config{Logger: logger.Default.LogMode(logger.Silent)},
	)
}

func (store *FileStore) KeyStore() (KeyStore, error) {
	if store.keyStore == nil {
		db, err := store.OpenDB("keys.db")
		if err != nil {
			return nil, err
		}

		store.keyStore, err = NewGormKeyStore(db)
		if err != nil {
			return nil, err
		}
	}

	return store.keyStore, nil
}
