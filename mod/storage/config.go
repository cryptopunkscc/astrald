package storage

type configSource struct {
	Exec string   `yaml:"exec"`
	Args []string `yaml:"args"`
}

type Config struct {
	Sources []configSource `yaml:"sources"`
}

var defaultConfig = Config{
	Sources: []configSource{},
}
