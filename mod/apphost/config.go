package apphost

type Config struct {
	// Listen on these adresses
	Listen []string `yaml:"listen"`

	// Runtime maps runtimes' names to paths of their executables.
	Runtime map[string]string `yaml:"runtime"`

	// Number of apphost workers
	Workers int `yaml:"workers"`

	// Allow anonymous sessions (with an empty access token)
	AllowAnonymous bool `yaml:"allow_anonymous"`
}

var defaultConfig = Config{
	Listen: []string{
		"tcp:127.0.0.1:8625",
		"unix:~/.apphost.sock",
	},
	Runtime:        map[string]string{},
	Workers:        256,
	AllowAnonymous: false,
}
