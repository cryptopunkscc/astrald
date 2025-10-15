package ip

const ModuleName = "ip"

type Module interface {
	LocalIPs() ([]IP, error)
	PublicIPCandidates() []IP
}

type PublicIPCandidateProvider interface {
	PublicIPCandidates() []IP // sorted
}
