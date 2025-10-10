package ip

const ModuleName = "ip"

type Module interface {
	// FIXME: renmame "config"
	ConfigIPs() (ips []IP)
	LocalIPs() (ips []IP, err error)
	FindIPCandidates() (ips []IP)
}

type CandidateFinder interface {
	FindIPCandidate() []IP // sorted
}
