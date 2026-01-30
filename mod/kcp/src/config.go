package kcp

import (
	"time"
)

type Config struct {
	Listen *bool `yaml:"listen,omitempty"`
	Dial   *bool `yaml:"dial,omitempty"`

	Endpoints   []string      `yaml:"configEndpoints,omitempty"`
	ListenPort  int           `yaml:"listen_port,omitempty"`
	DialTimeout time.Duration `yaml:"dial_timeout,omitempty"`
}

var defaultConfig = Config{
	DialTimeout: time.Minute,
	ListenPort:  1792,
}
