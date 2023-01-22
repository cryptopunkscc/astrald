package config

import "os"

type Config struct {
	Alias string `yaml:"alias"`
	Infra Infra  `yaml:"infra"`
	Log   Log    `yaml:"log"`
}

func (c Config) GetAlias() string {
	if c.Alias != "" {
		return c.Alias
	}

	if host, err := os.Hostname(); err == nil {
		return host
	}

	return ""
}
