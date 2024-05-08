package archives

import "github.com/cryptopunkscc/astrald/net"

type Config struct {
	AutoIndexZones string
}

var defaultConfig = Config{
	AutoIndexZones: (net.ZoneDevice | net.ZoneVirtual).String(),
}
