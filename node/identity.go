package node

import (
	"errors"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/node/storage"
	"log"
	"os"
)

const defaultIdentityKey = "id"

func setupIdentity(storage *storage.FilesystemStorage) id.Identity {
	// Try to load an existing identity
	bytes, err := storage.LoadBytes(defaultIdentityKey)
	if err == nil {
		identity, err := id.ParsePrivateKey(bytes)
		if err != nil {
			panic(err)
		}
		return identity
	}

	// The only acceptable error is ErrNotExist
	if !errors.Is(err, os.ErrNotExist) {
		panic(err)
	}

	// Generate a new identity
	log.Println("generating new node identity...")
	identity, err := id.GenerateIdentity()
	if err != nil {
		panic(err)
	}

	// Save the new identity
	err = storage.StoreBytes(defaultIdentityKey, identity.PrivateKey().Serialize())
	if err != nil {
		log.Println("error saving identity:", err)
	}

	return identity
}
