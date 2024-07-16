package tcp

import "time"

type Config struct {
	DialTimeout     time.Duration `yaml:"dial_timeout"`
	PublicEndpoints []string      `yaml:"public_endpoints"`
	ListenPort      int           `yaml:"listen_port"`
}

var defaultConfig = Config{
	DialTimeout: time.Minute,
	ListenPort:  1791,
}
