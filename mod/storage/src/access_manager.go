package storage

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/mod/storage"
	"github.com/cryptopunkscc/astrald/sig"
	"sync"
)

var _ storage.AccessManager = &AccessManager{}

type AccessManager struct {
	*Module
	verifiers sig.Set[storage.AccessVerifier]
	mu        sync.Mutex
}

func NewAccessManager(module *Module) *AccessManager {
	return &AccessManager{
		Module: module,
	}
}

func (mod *AccessManager) Verify(identity id.Identity, dataID data.ID) bool {
	// local node has access to everything
	if identity.IsEqual(mod.node.Identity()) {
		return true
	}

	for _, verifier := range mod.verifiers.Clone() {
		if verifier.Verify(identity, dataID) {
			return true
		}
	}

	return false
}

func (mod *AccessManager) AddAccessVerifier(verifier storage.AccessVerifier) {
	mod.verifiers.Add(verifier)
}

func (mod *AccessManager) RemoveAccessVerifier(verifier storage.AccessVerifier) {
	mod.verifiers.Remove(verifier)
}
