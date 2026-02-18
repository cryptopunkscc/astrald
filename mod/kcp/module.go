package kcp

import (
	"github.com/cryptopunkscc/astrald/mod/exonet"
)

const ModuleName = "kcp"

const (
	MethodNewEphemeralListener          = "kcp.new_ephemeral_listener"
	MethodCloseEphemeralListener        = "kcp.close_ephemeral_listener"
	MethodAddRemoteEndpointLocalPort    = "kcp.add_remote_endpoint_local_port"
	MethodRemoveRemoteEndpointLocalPort = "kcp.remove_remote_endpoint_local_port"
	MethodListEndpointsLocalMappings    = "kcp.list_endpoints_local_mappings"
)

type Module interface {
	exonet.Dialer
	exonet.Unpacker
	exonet.Parser
	ListenPort() int
}
