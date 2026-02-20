package ip

import (
	"github.com/cryptopunkscc/astrald/mod/ip"
)

func (mod *Module) AddPublicIPCandidateProvider(provider ip.PublicIPCandidateProvider) error {
	return mod.providers.Add(provider)
}

func (mod *Module) PublicIPCandidates() (ips []ip.IP) {
	var unique = map[string]struct{}{}

	for _, provider := range mod.providers.Clone() {
		for _, i := range provider.PublicIPCandidates() {
			if _, found := unique[i.String()]; found {
				continue
			}

			unique[i.String()] = struct{}{}
			ips = append(ips, i)
		}
	}

	return
}
