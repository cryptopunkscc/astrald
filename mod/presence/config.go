package presence

type Config struct {
	// Make the node discoverable in the local network
	Discoverable bool `yaml:"discoverable"`

	// Automatically add present nodes to contacts
	AutoAdd bool `yaml:"auto_add"`

	// Trust self-assigned aliases
	TrustAliases bool `yaml:"trust_aliases"`
}

var defaultConfig = Config{
	Discoverable: true,
	AutoAdd:      true,
	TrustAliases: true,
}
