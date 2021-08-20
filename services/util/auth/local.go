package auth

import "github.com/cryptopunkscc/astrald/api"

func Local(core api.Core, conn api.ConnectionRequest) bool {
	if conn.Caller() != core.Network().Identity() {
		conn.Reject()
		return false
	}
	return true
}
