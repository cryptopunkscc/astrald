package tcp

import "time"

type Config struct {
	DialTimeout     time.Duration `yaml:"dial_timeout,omitempty"`
	PublicEndpoints []string      `yaml:"public_endpoints,omitempty"`
	ListenPort      int           `yaml:"listen_port,omitempty"`
}

var defaultConfig = Config{
	DialTimeout: time.Minute,
	ListenPort:  1791,
}
