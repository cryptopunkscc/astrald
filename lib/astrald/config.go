package astrald

import (
	"os"

	"github.com/cryptopunkscc/astrald/lib/apphost"
)

type Config struct {
	Endpoint string
	Token    string
}

func DefaultConfig() Config {
	return Config{
		Endpoint: apphost.DefaultEndpoint,
		Token:    os.Getenv(apphost.AuthTokenEnv),
	}
}
