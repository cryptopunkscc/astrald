package fwd

type Config struct {
	Forwards map[string]string `yaml:"forwards"`
}

var defaultConfig = Config{
	Forwards: map[string]string{},
}
