package apphost

type Config struct {
	// Listen on these addresses
	Listen []string `yaml:"listen,omitempty"`

	// Number of apphost workers
	Workers int `yaml:"workers,omitempty"`

	Tokens map[string]string `yaml:"tokens,omitempty"`

	BindHTTP string `yaml:"bind_http,flow"`

	AllowAnonymous bool `yaml:"allow_anonymous,omitempty"`

	// TrustedWebOrigins are the browser origins allowed to act anonymously over
	// the WS transport (route local-zone queries and request a token). Exact
	// match against the request Origin header. Add a dev-server origin here to
	// develop the settings app against this node.
	TrustedWebOrigins []string `yaml:"trusted_web_origins,omitempty"`
}

var defaultConfig = Config{
	Listen: []string{
		"tcp:127.0.0.1:8625",
		"unix:~/.apphost.sock",
		"memu:apphosty",
		"memb:apphostb",
	},
	BindHTTP:          "tcp:0.0.0.0:8624",
	Tokens:            map[string]string{},
	Workers:           32,
	AllowAnonymous:    true,
	TrustedWebOrigins: []string{TrustedWebOrigin},
}
