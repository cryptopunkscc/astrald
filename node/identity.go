package node

import (
	"errors"
	"github.com/cryptopunkscc/astrald/auth/id"
	"log"
	"os"
)

const defaultIdentityKey = "id"

func (node *Node) setupIdentity() error {
	var err error

	// Try to load an existing identity
	bytes, err := node.Store.LoadBytes(defaultIdentityKey)
	if err == nil {
		node.identity, err = id.ParsePrivateKey(bytes)
		return err
	}

	if !errors.Is(err, os.ErrNotExist) {
		return err
	}

	// Generate a new identity
	log.Println("generating new node identity...")
	node.identity, err = id.GenerateIdentity()
	if err != nil {
		return err
	}

	// Save the new identity
	err = node.Store.StoreBytes(defaultIdentityKey, node.identity.PrivateKey().Serialize())
	if err != nil {
		return err
	}

	return nil
}
