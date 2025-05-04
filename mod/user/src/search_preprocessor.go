package user

import "github.com/cryptopunkscc/astrald/mod/objects"

func (mod *Module) PreprocessSearch(search *objects.Search) {
	ac := mod.ActiveContract()
	if ac == nil {
		return
	}

	if !search.CallerID.IsEqual(ac.UserID) {
		return
	}

	for _, nodeID := range mod.getLinkedSibs() {
		search.Sources = append(search.Sources, nodeID)
	}
}
