package ip

import "errors"

const ModuleName = "ip"

// Module is the public API of the ip module; LocalIPs returns addresses bound to local interfaces,
// while PublicIPCandidates returns the subset considered reachable from the internet.
type Module interface {
	LocalIPs() ([]IP, error)
	PublicIPCandidates() []IP
	DefaultGateway() (IP, error)
}

type PublicIPCandidateProvider interface {
	PublicIPCandidates() []IP // sorted
}

var ErrDefaultGatewayNotFound = errors.New("default gateway not found")
