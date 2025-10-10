package udp

import (
	"time"

	"github.com/cryptopunkscc/astrald/mod/udp/rudp"
)

type Config struct {
	ListenPort      int           `yaml:"listen_port,omitempty"` // Port to listen on for incoming connections (default 1791)
	PublicEndpoints []string      `yaml:"public_endpoints,omitempty"`
	DialTimeout     time.Duration `yaml:"dial_timeout,omitempty"`     // Timeout for dialing connections (default 1 minute)
	TransportConfig rudp.Config   `yaml:"transport_config,omitempty"` // Flow control settings for UDP connections
}

var defaultConfig = Config{
	DialTimeout: time.Minute,
	ListenPort:  1791,
}
