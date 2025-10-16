package ip

import "errors"

const ModuleName = "ip"

type Module interface {
	LocalIPs() ([]IP, error)
	PublicIPCandidates() []IP
	DefaultGateway() (IP, error)
}

type PublicIPCandidateProvider interface {
	PublicIPCandidates() []IP // sorted
}

var ErrDefaultGatewayNotFound = errors.New("default gateway not found")
