package storage

type Config struct {
	LocalFiles []string `yaml:"local_files"`
}

var defaultConfig = Config{}
