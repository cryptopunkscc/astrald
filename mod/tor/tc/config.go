package tc

const defaultControlAddr = "127.0.0.1:9051"

type Config struct {
	ControlAddr string
}

func (cfg Config) getContolAddr() string {
	if cfg.ControlAddr != "" {
		return cfg.ControlAddr
	}
	return defaultControlAddr
}
