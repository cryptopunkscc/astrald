package apphost

type Config struct {
	// Listen on these adresses
	Listen []string `yaml:"listen,omitempty"`

	// Number of apphost workers
	Workers int `yaml:"workers,omitempty"`

	Tokens map[string]string `yaml:"tokens,omitempty"`

	BindHTTP string `yaml:"bind_http,flow"`

	AllowAnonymous bool `yaml:"allow_anonymous,omitempty"`
}

var defaultConfig = Config{
	Listen: []string{
		"tcp:127.0.0.1:8625",
		"unix:~/.apphost.sock",
		"memu:apphosty",
		"memb:apphostb",
	},
	BindHTTP:       "tcp:0.0.0.0:8624",
	Tokens:         map[string]string{},
	Workers:        32,
	AllowAnonymous: true,
}
