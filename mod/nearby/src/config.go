package nearby

import "github.com/cryptopunkscc/astrald/mod/nearby"

const aliasPrefix = "."

type Config struct {
	Mode nearby.Mode `yaml:"mode"`
}

var defaultConfig = Config{
	Mode: nearby.ModeStealth,
}
