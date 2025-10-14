package ip

import (
	"github.com/cryptopunkscc/astrald/mod/ip"
)

func (mod *Module) AddFinder(finder ip.CandidateFinder) error {
	return mod.finders.Add(finder)
}

func (mod *Module) FindIPCandidates() (ips []ip.IP) {
	var unique = map[string]struct{}{}

	for _, finder := range mod.finders.Clone() {
		for _, i := range finder.FindIPCandidate() {
			if _, found := unique[i.String()]; found {
				continue
			}
			unique[i.String()] = struct{}{}
			ips = append(ips, i)
		}
	}

	return
}
