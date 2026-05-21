package apphost

type Config struct {
	// Listen on these adresses
	Listen []string `yaml:"listen,omitempty"`

	// Number of apphost workers
	Workers int `yaml:"workers,omitempty"`

	Tokens map[string]string `yaml:"tokens,omitempty"`

	BindHTTP string `yaml:"bind_http,flow"`

	AllowAnonymous bool `yaml:"allow_anonymous,omitempty"`

	// WSAllowOrigins lists origins permitted for the /.ws WebSocket endpoint
	// (host patterns matched by path.Match, see coder/websocket AcceptOptions).
	// Empty means loopback-only — non-loopback requests are refused.
	WSAllowOrigins []string `yaml:"ws_allow_origins,omitempty"`
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
