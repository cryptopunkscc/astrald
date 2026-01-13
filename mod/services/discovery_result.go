package services

type ServiceDiscoveryResultKind uint8

const (
	DiscoveryEventChange ServiceDiscoveryResultKind = iota
	DiscoveryEventFlush
)

type ServiceDiscoveryResult struct {
	Kind   ServiceDiscoveryResultKind
	Change ServiceChange
}

func DiscoveryChange(change ServiceChange) ServiceDiscoveryResult {
	return ServiceDiscoveryResult{Kind: DiscoveryEventChange, Change: change}
}

func DiscoveryFlush() ServiceDiscoveryResult {
	return ServiceDiscoveryResult{Kind: DiscoveryEventFlush}
}

func IsDiscoveryFlush(ev ServiceDiscoveryResult) bool {
	return ev.Kind == DiscoveryEventFlush
}
