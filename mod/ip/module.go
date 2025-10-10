package ip

const ModuleName = "ip"

type Module interface {
	ConfiguredIPs() (ips []IP)
	LocalIPs() (ips []IP, err error)
	FindIPCandidates() (ips []IP)
}

type CandidateFinder interface {
	FindIPCandidate() []IP // sorted
}
