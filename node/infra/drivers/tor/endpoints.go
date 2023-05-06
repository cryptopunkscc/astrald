package tor

import "github.com/cryptopunkscc/astrald/net"

func (drv *Driver) Endpoints() []net.Endpoint {
	if drv.serviceAddr.IsZero() {
		return nil
	}
	return []net.Endpoint{drv.serviceAddr}
}
