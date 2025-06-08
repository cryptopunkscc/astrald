package nearby

const aliasPrefix = "."

type Config struct {
	// Make the node discoverable in the local network
	Visible bool `yaml:"visible"`

	// Automatically add present nodes to contacts
	AutoAdd bool `yaml:"auto_add"`

	// Trust self-assigned aliases
	TrustAliases bool `yaml:"trust_aliases"`
}

var defaultConfig = Config{
	Visible:      true,
	AutoAdd:      true,
	TrustAliases: true,
}
