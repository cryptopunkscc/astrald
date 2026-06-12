package user

import "github.com/cryptopunkscc/astrald/mod/objects"

// PreprocessSearch injects linked siblings as additional search sources when the caller is the active contract issuer.
func (mod *Module) PreprocessSearch(search *objects.Search) {
	ac := mod.ActiveContract()
	if ac == nil {
		return
	}

	if !search.CallerID.IsEqual(ac.Issuer) {
		return
	}

	for _, nodeID := range mod.getLinkedSibs() {
		search.Sources = append(search.Sources, nodeID)
	}
}
