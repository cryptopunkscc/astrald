package tor

import "time"

const (
	defaultTorProxy    = "127.0.0.1:9050"
	defaultControlAddr = "127.0.0.1:9051"
	defaultDialTimeout = time.Minute
	defaultListenPort  = 1791
	defaultBackend     = "system"
)

type Config struct {
	Backend     string `yaml:"backend"`
	TorProxy    string `yaml:"tor_proxy"`
	ControlAddr string `yaml:"control_addr"`
	DialTimeout string `yaml:"dial_timeout"`
	ListenPort  int
}

func (cfg Config) GetBackend() string {
	if cfg.Backend == "" {
		if _, found := backends[defaultBackend]; found {
			return defaultBackend
		}
		for k, _ := range backends {
			return k
		}
	}
	return cfg.Backend
}

func (cfg Config) GetListenPort() int {
	if cfg.ListenPort != 0 {
		return cfg.ListenPort
	}
	return defaultListenPort
}

func (cfg Config) GetContolAddr() string {
	if cfg.ControlAddr != "" {
		return cfg.ControlAddr
	}
	return defaultControlAddr
}

func (cfg Config) GetProxyAddress() string {
	if cfg.TorProxy != "" {
		return cfg.TorProxy
	}
	return defaultTorProxy
}

func (cfg Config) GetDialTimeout() time.Duration {
	timeout, err := time.ParseDuration(cfg.DialTimeout)
	if err != nil {
		return defaultDialTimeout
	}
	return timeout
}
