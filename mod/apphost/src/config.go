package apphost

type Config struct {
	// Listen on these adresses
	Listen []string `yaml:"listen"`

	// Number of apphost workers
	Workers int `yaml:"workers"`

	// Identity to use for anonymous connections
	DefaultIdentity string `yaml:"default_identity"`

	Tokens  map[string]string `yaml:"tokens"`
	Autorun []configRun       `yaml:"autorun"`

	RoutePriority int `yaml:"route_priority"`
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
		"memu:apphost",
		"memb:apphost",
	},
	Tokens:        map[string]string{},
	Workers:       256,
	RoutePriority: 90,
}
