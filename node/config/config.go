package config

import "os"

const configKey = "astrald.conf"

type Config struct {
	Alias string `yaml:"alias"`
	Infra Infra  `yaml:"infra"`
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
