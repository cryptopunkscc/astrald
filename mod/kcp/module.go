package kcp

import (
	"github.com/cryptopunkscc/astrald/mod/exonet"
)

const ModuleName = "kcp"

const (
	MethodNewEphemeralListener      = "kcp.new_ephemeral_listener"
	MethodCloseEphemeralListener    = "kcp.close_ephemeral_listener"
	MethodSetEndpointLocalPort      = "kcp.set_endpoint_local_port"
	MethodRemoveEndpointLocalPort   = "kcp.remove_endpoint_local_port"
	MethodListEndpointLocalMappings = "kcp.list_endpoint_local_mappings"
)

// Module is the public contract for the KCP transport: it dials, unpacks, and parses
// exonet endpoints over KCP/UDP and exposes the port it is bound to.
type Module interface {
	exonet.Dialer
	exonet.Unpacker
	exonet.Parser
	// ListenPort returns the UDP port the module is currently bound to; 0 means unbound.
	ListenPort() int
}
