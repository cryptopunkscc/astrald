package node

import (
	"errors"
	_id "github.com/cryptopunkscc/astrald/node/auth/id"
	_fs "github.com/cryptopunkscc/astrald/node/fs"
	"log"
	"os"
)

const defaultIdentityFilename = "id"

func setupIdentity(fs *_fs.Filesystem) *_id.ECIdentity {
	// Try to load an existing identity
	idBytes, err := fs.Read(defaultIdentityFilename)
	if err == nil {
		id, err := _id.ECIdentityFromBytes(idBytes)
		if err != nil {
			panic(err)
		}
		return id
	}

	// The only acceptable error is ErrNotExist
	if !errors.Is(err, os.ErrNotExist) {
		panic(err)
	}

	// Generate a new identity
	log.Println("generating new node identity...")
	id, err := _id.GenerateECIdentity()
	if err != nil {
		panic(err)
	}

	// Save the new identity
	err = fs.Write(defaultIdentityFilename, id.PrivateKey().Serialize())
	if err != nil {
		log.Println("error saving identity:", err)
	}

	return id
}
