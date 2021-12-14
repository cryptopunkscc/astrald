package tor

import "time"

const defaultTorProxy = "127.0.0.1:9050"
const defaultControlAddr = "127.0.0.1:9051"
const defaultDialTimeout = time.Minute
const defaultListenPort = 1791

type Config struct {
	TorProxy    string `yaml:"tor_proxy"`
	ControlAddr string `yaml:"control_addr"`
	DialTimeout string `yaml:"dial_timeout"`
	DataDir     string `yaml:"data_dir"`
	ListenPort  int
	cont        string
}

func (cfg Config) getListenPort() int {
	if cfg.ListenPort != 0 {
		return cfg.ListenPort
	}
	return defaultListenPort
}

func (cfg Config) getContolAddr() string {
	if cfg.ControlAddr != "" {
		return cfg.ControlAddr
	}
	return defaultControlAddr
}

func (cfg Config) getProxyAddress() string {
	if cfg.TorProxy != "" {
		return cfg.TorProxy
	}
	return defaultTorProxy
}

func (cfg Config) getDialTimeout() time.Duration {
	timeout, err := time.ParseDuration(cfg.DialTimeout)
	if err != nil {
		return defaultDialTimeout
	}
	return timeout
}
