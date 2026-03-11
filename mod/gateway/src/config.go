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
	Gateway  GatewayConfig      `yaml:"gateway"`
	Gateways []*astral.Identity `yaml:"gateways"`

	Visibility gateway.Visibility `yaml:"visibility"`
	InitConns  int32              `yaml:"init_conns"`
	MaxConns   int32              `yaml:"max_conns"`
}

var defaultConfig = Config{
	Visibility: gateway.VisibilityPublic,
	InitConns:  1,
	MaxConns:   8,
}
