package gateway

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/gateway"
)

const defaultGateway = "node1f3AwbE1gJAgAqEx98FMipokcaE9ZapIphzDUkAceE7Pmw8ghmFV19QKCATeC7uyoLszQA"

type GatewayConfig struct {
	Enabled bool     `yaml:"enabled"`
	Listen  []string `yaml:"listen"`
}

type Config struct {
	Gateway    GatewayConfig      `yaml:"gateway"`
	Visibility gateway.Visibility `yaml:"visibility"`
	Gateways   []*astral.Identity `yaml:"gateways"`
}

var defaultConfig = Config{
	Visibility: gateway.VisibilityPublic,
}
