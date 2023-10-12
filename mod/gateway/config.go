package gateway

type Config struct {
	Subscribe []string `yaml:"subscribe"`
}

var defaultConfig = Config{}
