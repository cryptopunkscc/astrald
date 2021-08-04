package node

import (
	"errors"
	_fs "github.com/cryptopunkscc/astrald/node/fs"
	"gopkg.in/yaml.v2"
	"log"
	"os"
)

const defaultPort = 1791
const defaultConfigFilename = "astrald.conf"

type Config struct {
	Port int
}

var defaultConfig = Config{
	Port: defaultPort,
}

func loadConfig(fs *_fs.Filesystem) *Config {
	var cfg = defaultConfig

	// Load the config file
	configBytes, err := fs.Read(defaultConfigFilename)
	if err == nil {
		// Parse config file
		err = yaml.Unmarshal(configBytes, &cfg)
		if err != nil {
			log.Println("error parsing config file:", err)
		}
	} else {
		if !errors.Is(err, os.ErrNotExist) {
			log.Println("error reading config file:", err)
		}
	}

	return &cfg
}
