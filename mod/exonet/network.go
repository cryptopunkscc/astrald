package exonet

import "github.com/cryptopunkscc/astrald/net"

// Network returns link's network name or unknown if network could not be determined
func Network(link net.Link) string {
	var t = link.Transport().(Conn)
	if t == nil {
		return "unknown"
	}

	if e := t.RemoteEndpoint(); e != nil {
		return e.Network()
	}
	if e := t.LocalEndpoint(); e != nil {
		return e.Network()
	}

	return "unknown"
}