package user

type Config struct {
	Identity string `yaml:"identity"`
}

var defaultConfig = Config{}
