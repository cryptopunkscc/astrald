package nodes

type Config struct {
	LogPings bool `yaml:"log_pings"`
}

var defaultConfig = Config{}
