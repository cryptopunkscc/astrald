package kcp

import (
	"time"
)

type Config struct {
	Endpoints   []string      `yaml:"configEndpoints,omitempty"`
	ListenPort  int           `yaml:"listen_port,omitempty"`
	DialTimeout time.Duration `yaml:"dial_timeout,omitempty"`
}

var defaultConfig = Config{
	DialTimeout: time.Minute,
	ListenPort:  1792,
}
