package gateway

import "github.com/cryptopunkscc/astrald/mod/gateway"

const defaultGateway = "node1f3AwbE1gJAgAqEx98FMipokcaE9ZapIphzDUkAceE7Pmw8ghmFV19QKCATeC7uyoLszQA"

const (
	defaultInitConns = 1
	defaultMaxConns  = 8
)

type Config struct {
	ActAsGateway bool `yaml:"act_as_gateway"`

	Sockets    map[string]uint16  `yaml:"sockets"`
	Visibility gateway.Visibility `yaml:"visibility"`
	Gateways   []string           `yaml:"gateways"`
	InitConns  int32              `yaml:"init_conns"`
	MaxConns   int32              `yaml:"max_conns"`
}

var defaultConfig = Config{
	Visibility: gateway.VisibilityPublic,
	Gateways: []string{
		defaultGateway,
	},
	InitConns: defaultInitConns,
	MaxConns:  defaultMaxConns,
}
