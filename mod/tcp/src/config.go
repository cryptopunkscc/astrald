package tcp

import "time"

type Config struct {
	Listen *bool `yaml:"listen,omitempty"`
	Dial   *bool `yaml:"dial,omitempty"`

	Endpoints   []string      `yaml:"configEndpoints,omitempty"`
	DialTimeout time.Duration `yaml:"dial_timeout,omitempty"`
	ListenPort  int           `yaml:"listen_port,omitempty"`
}

var trueVal = true
var defaultConfig = Config{
	Dial:        &trueVal,
	Listen:      &trueVal,
	DialTimeout: time.Minute,
	ListenPort:  1791,
}
