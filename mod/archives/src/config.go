package archives

import "github.com/cryptopunkscc/astrald/astral"

type Config struct {
	AutoIndexZones string
}

var defaultConfig = Config{
	AutoIndexZones: (astral.ZoneDevice | astral.ZoneVirtual).String(),
}
