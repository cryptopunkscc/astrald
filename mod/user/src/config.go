package user

type Config struct {
	LocalUser string `yaml:"local_user"`
}

var defaultConfig = Config{}
