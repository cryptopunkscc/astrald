package ether

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/tcp"
)

const ModuleName = "ether"

// Module provides public communication within various broadcast networks such as local area networks.
type Module interface {
	// Push an object to the ether
	Push(astral.Object, *astral.Identity) error
	PushToIP(tcp.IP, astral.Object, *astral.Identity) error
}
