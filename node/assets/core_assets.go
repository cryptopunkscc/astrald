package assets

import (
	"errors"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/resources"
	"gopkg.in/yaml.v2"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"path/filepath"
	"strings"
	"time"
)

var _ Assets = &CoreAssets{}

var dbOpen func(string) gorm.Dialector

type CoreAssets struct {
	res    resources.Resources
	log    *log.Logger
	prefix string
}

func NewCoreAssets(res resources.Resources, log *log.Logger) *CoreAssets {
	return &CoreAssets{res: res, log: log}
}

func (assets *CoreAssets) Res() resources.Resources {
	return assets.res
}

func (assets *CoreAssets) Read(name string) ([]byte, error) {
	return assets.res.Read(assets.prefix + name)
}

func (assets *CoreAssets) Write(name string, data []byte) error {
	return assets.res.Write(assets.prefix+name, data)
}

func (assets *CoreAssets) LoadYAML(name string, out interface{}) error {
	if !strings.HasSuffix(strings.ToLower(name), ".yaml") {
		name = name + ".yaml"
	}

	bytes, err := assets.Res().Read(assets.prefix + name)
	if err != nil {
		return err
	}

	return yaml.Unmarshal(bytes, out)
}

func (assets *CoreAssets) StoreYAML(name string, in interface{}) error {
	if !strings.HasSuffix(strings.ToLower(name), ".yaml") {
		name = name + ".yaml"
	}

	bytes, err := yaml.Marshal(in)
	if err != nil {
		return err
	}

	return assets.Res().Write(assets.prefix+name, bytes)
}

func (assets *CoreAssets) OpenDB(name string) (*gorm.DB, error) {
	if name == "" {
		return nil, errors.New("invalid name")
	}
	if !strings.HasSuffix(name, ".db") {
		name = name + ".db"
	}

	var l logger.Interface
	if assets.log != nil {
		l = logger.New(
			&logWriter{Logger: assets.log},
			logger.Config{
				SlowThreshold:             200 * time.Millisecond,
				LogLevel:                  logger.Warn,
				IgnoreRecordNotFoundError: false,
				Colorful:                  true,
			})
	} else {
		l = logger.Default.LogMode(logger.Silent)
	}

	var cfg = &gorm.Config{
		Logger: l,
	}

	switch res := assets.res.(type) {
	case *resources.FileResources:
		var dbPath = filepath.Join(res.Root(), assets.prefix+name)

		return gorm.Open(
			dbOpen(dbPath),
			cfg,
		)

	case *resources.MemResources:
		var dbPath = "file::memory:?cache=shared"

		return gorm.Open(
			dbOpen(dbPath),
			cfg,
		)
	}

	return nil, errors.New("database unavailable")
}

func (assets *CoreAssets) WithPrefix(prefix string) *CoreAssets {
	return &CoreAssets{
		res:    assets.res,
		log:    assets.log,
		prefix: prefix,
	}
}

type logWriter struct {
	*log.Logger
	level int
}

func (w *logWriter) Printf(s string, i ...interface{}) {
	w.Logger.Logv(w.level, s, i...)
}
