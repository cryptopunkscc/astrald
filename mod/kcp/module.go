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

type Module interface {
	exonet.Dialer
	exonet.Unpacker
	exonet.Parser
	ListenPort() int
}
