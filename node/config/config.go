package config

import "os"

const configKey = "astrald.conf"

type Config struct {
	Alias     string   `yaml:"alias"`
	Infra     Infra    `yaml:"infra"`
	LogEvents []string `yaml:"log_events"`
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

func (c Config) LogEventsInclude(s string) bool {
	if s == "" {
		return false
	}
	if len(c.LogEvents) == 0 {
		return false
	}
	for _, e := range c.LogEvents {
		if e == s {
			return true
		}
	}
	return false
}
