package tcp

import "time"

type Config struct {
	Endpoints   []string      `yaml:"configEndpoints,omitempty"`
	DialTimeout time.Duration `yaml:"dial_timeout,omitempty"`
	ListenPort  int           `yaml:"listen_port,omitempty"`
}

var defaultConfig = Config{
	DialTimeout: time.Minute,
	ListenPort:  1791,
}
