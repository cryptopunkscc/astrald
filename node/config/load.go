package config

import (
	"errors"
	"github.com/cryptopunkscc/astrald/storage"
	"gopkg.in/yaml.v2"
	"os"
)

func Load(store storage.Store) (Config, error) {
	var cfg = defaultConfig

	// Load the config file
	bytes, err := store.LoadBytes(configKey)
	if err == nil {
		// Parse config file
		err = yaml.Unmarshal(bytes, &cfg)
		if err != nil {
			return defaultConfig, err
		}
	} else {
		if !errors.Is(err, os.ErrNotExist) {
			return defaultConfig, err
		}
	}

	return cfg, nil
}
