package node

import (
	"errors"
	iastral "github.com/cryptopunkscc/astrald/infra/astral"
	"github.com/cryptopunkscc/astrald/infra/inet"
	"github.com/cryptopunkscc/astrald/infra/tor"
	"github.com/cryptopunkscc/astrald/node/storage"
	"gopkg.in/yaml.v2"
	"log"
	"os"
)

const configKey = "astrald.conf"

type Config struct {
	Alias string `yaml:"alias"`

	Net struct {
		Inet   inet.Config    `yaml:"inet"`
		Tor    tor.Config     `yaml:"tor"`
		Astral iastral.Config `yaml:"astral"`
	} `yaml:"net"`
}

var defaultConfig = Config{}

func loadConfig(store storage.Store) *Config {
	var cfg = defaultConfig

	// Load the config file
	configBytes, err := store.LoadBytes(configKey)
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
