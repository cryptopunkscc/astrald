package node

import (
	"errors"
	"github.com/cryptopunkscc/astrald/node/network"
	"github.com/cryptopunkscc/astrald/node/storage"
	"gopkg.in/yaml.v2"
	"log"
	"os"
)

const defaultConfigFilename = "astrald.conf"

type Config struct {
	Network network.Config
	Alias   map[string]string
}

var defaultConfig = Config{}

func loadConfig(store storage.Store) *Config {
	var cfg = defaultConfig

	// Load the config file
	configBytes, err := store.LoadBytes(defaultConfigFilename)
	if err == nil {
		// Parse config file
		err = yaml.Unmarshal(configBytes, &cfg)
		if err != nil {
			log.Println("error parsing config:", err)
		}
	} else {
		if !errors.Is(err, os.ErrNotExist) {
			log.Println("error reading config:", err)
		}
	}

	return &cfg
}

func (config Config) getAlias(alias string) string {
	return config.Alias[alias]
}
