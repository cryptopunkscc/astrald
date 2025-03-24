package apphost

type ObjectServerConfig struct {
	Bind []string `yaml:"bind,omitempty"`
}

type Config struct {
	// Listen on these adresses
	Listen []string `yaml:"listen,omitempty"`

	// Number of apphost workers
	Workers int `yaml:"workers,omitempty"`

	Tokens map[string]string `yaml:"tokens,omitempty"`

	ObjectServer ObjectServerConfig `yaml:"object_server,omitempty"`

	AllowAnonymous bool `yaml:"allow_anonymous,omitempty"`
}

var defaultConfig = Config{
	Listen: []string{
		"tcp:127.0.0.1:8625",
		"unix:~/.apphost.sock",
		"memu:apphosty",
		"memb:apphostb",
	},
	ObjectServer: ObjectServerConfig{
		Bind: []string{
			"tcp:127.0.0.1:8624",
		},
	},
	Tokens:         map[string]string{},
	Workers:        32,
	AllowAnonymous: true,
}
