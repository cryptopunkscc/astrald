package node

import (
	"github.com/cryptopunkscc/astrald/node/config"
	"os"
)

const configName = "node"

type Config struct {
	store config.Store
	data  struct {
		Alias string `yaml:"alias,omitempty"`
	}
}

type LogConfig struct {
	IncludeEvents []string          `yaml:"include_events,omitempty"`
	ExcludeEvents []string          `yaml:"exclude_events,omitempty"`
	Level         int               `yaml:"level,omitempty"`
	TagLevels     map[string]int    `yaml:"tag_levels,omitempty"`
	TagColors     map[string]string `yaml:"tag_colors,omitempty"`
	HideDate      bool              `yaml:"hide_date,omitempty"`
}

func LoadConfig(store config.Store) (*Config, error) {
	var cfg = defaultConfig(store)

	return cfg, store.LoadYAML(configName, &cfg.data)
}

func (c *Config) Alias() string {
	return c.data.Alias
}

func (c *Config) SetAlias(alias string) error {
	c.data.Alias = alias

	return c.save()
}

func (c *Config) save() error {
	return c.store.StoreYAML(configName, c.data)
}

func (c *LogConfig) isIncluded(event string) bool {
	if len(c.IncludeEvents) == 0 {
		return false
	}
	for _, e := range c.IncludeEvents {
		if e == event {
			return true
		}
	}
	return false
}

func (c *LogConfig) isExcluded(event string) bool {
	if len(c.ExcludeEvents) == 0 {
		return false
	}
	for _, e := range c.ExcludeEvents {
		if e == event {
			return true
		}
	}
	return false
}

func (c *LogConfig) IsEventLoggable(event string) bool {
	if len(c.ExcludeEvents) > 0 {
		return !c.isExcluded(event)
	}

	return c.isIncluded(event)
}

func defaultConfig(store config.Store) *Config {
	var cfg = &Config{store: store}

	if host, err := os.Hostname(); err == nil {
		cfg.data.Alias = host
	}

	return cfg
}
