package tor

import "time"

type Config struct {
	Listen *bool `yaml:"listen,omitempty"`
	Dial   *bool `yaml:"dial,omitempty"`

	TorProxy    string        `yaml:"tor_proxy"`
	ControlAddr string        `yaml:"control_addr"`
	DialTimeout time.Duration `yaml:"dial_timeout"`
	ListenPort  int
}

var defaultConfig = Config{
	TorProxy:    "127.0.0.1:9050",
	ControlAddr: "127.0.0.1:9051",
	DialTimeout: time.Minute,
	ListenPort:  1791,
}
