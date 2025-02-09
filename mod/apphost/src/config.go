package apphost

type ObjectServerConfig struct {
	Bind []string `yaml:"bind"`
}

type Config struct {
	// Listen on these adresses
	Listen []string `yaml:"listen"`

	// Number of apphost workers
	Workers int `yaml:"workers"`

	Tokens map[string]string `yaml:"tokens"`

	ObjectServer ObjectServerConfig `yaml:"object_server"`
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
	Tokens:  map[string]string{},
	Workers: 32,
}
