package config

import (
	"errors"
	"gopkg.in/yaml.v2"
	"os"
)

func LoadYAMLFile(filePath string) (Config, error) {
	var cfg = defaultConfig

	// Load the config file
	bytes, err := os.ReadFile(filePath)
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
