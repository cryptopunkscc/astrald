package apphost

type Config struct {
	// Listen on these adresses
	Listen []string `yaml:"listen"`

	// Number of apphost workers
	Workers int `yaml:"workers"`

	// Allow anonymous sessions (with an empty access token)
	AllowAnonymous bool `yaml:"allow_anonymous"`

	Tokens  map[string]string `yaml:"tokens"`
	Autorun []configRun       `yaml:"autorun"`
}

type configRun struct {
	Exec     string   `yaml:"exec"`
	Args     []string `yaml:"args"`
	Identity string   `yaml:"identity"`
}

var defaultConfig = Config{
	Listen: []string{
		"tcp:127.0.0.1:8625",
		"unix:~/.apphost.sock",
	},
	Tokens:         map[string]string{},
	Workers:        256,
	AllowAnonymous: false,
}
