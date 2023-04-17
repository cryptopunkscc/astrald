package node

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"os"
	"path/filepath"
)

const defaultIdentityKey = "id"

func (node *CoreNode) loadIdentity(name string) (id.Identity, error) {
	filePath := filepath.Join(node.rootDir, name)

	bytes, err := os.ReadFile(filePath)
	if err != nil {
		return id.Identity{}, err
	}

	return id.ParsePrivateKey(bytes)
}

func (node *CoreNode) generateIdentity(name string) (id.Identity, error) {
	filePath := filepath.Join(node.rootDir, name)

	identity, err := id.GenerateIdentity()
	if err != nil {
		return id.Identity{}, err
	}

	return identity, os.WriteFile(filePath, identity.PrivateKey().Serialize(), 0600)
}

func (node *CoreNode) setupIdentity() (err error) {
	if node.identity, err = node.loadIdentity(defaultIdentityKey); err == nil {
		return
	}

	node.identity, err = node.generateIdentity(defaultIdentityKey)
	return
}
