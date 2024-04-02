package media

type Config struct {
	AutoIndexNet []string `yaml:"auto_index_net"`
}

var defaultConfig = Config{
	AutoIndexNet: []string{
		"image/jpeg",
		"image/png",
		//"audio/mpeg",
	},
}
