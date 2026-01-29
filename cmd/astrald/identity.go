package main

import (
	"bytes"
	"fmt"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/crypto"
	"github.com/cryptopunkscc/astrald/mod/keys"
	"github.com/cryptopunkscc/astrald/mod/secp256k1"
	"github.com/cryptopunkscc/astrald/resources"
)

// loadNodeIdentity loads node's identity from resources. Generates a new identity if we don't have one yet.
func loadNodeIdentity(resources resources.Resources) (identity *astral.Identity, err error) {
	data, err := resources.Read(resNodeIdentity)
	if err == nil {
		object, _, _ := astral.Decode(bytes.NewReader(data), astral.Canonical())

		switch object := object.(type) {
		case *crypto.PrivateKey:
			fmt.Println("found crypto.PrivateKey")
			if object.Type != secp256k1.KeyType {
				return nil, fmt.Errorf("unsupported private key type: %s", object.Type)
			}

			identity, err = astral.IdentityFromPrivKeyBytes(object.Key)
			if err != nil {
				return nil, err
			}

		case *astral.PrivateIdentity:
			fmt.Println("found PrivateIdentity")

			identity = (*astral.Identity)(object)

		case *keys.PrivateKey:
			fmt.Println("found keys.PrivateKey")

			identity, err = astral.IdentityFromPrivKeyBytes(object.Bytes)
			if err != nil {
				return nil, err
			}

		default:
			return nil, fmt.Errorf("unknown identity object: %s", object.ObjectType())
		}
	} else {
		identity = astral.GenerateIdentity()
	}

	nodeKey := &crypto.PrivateKey{
		Type: secp256k1.KeyType,
		Key:  identity.PrivateKey().Serialize(),
	}

	// store node key
	var keyBytes = &bytes.Buffer{}
	_, err = astral.Encode(keyBytes, nodeKey, astral.Canonical())
	if err != nil {
		return nil, err
	}

	err = resources.Write("node_key", keyBytes.Bytes())
	if err != nil {
		return nil, err
	}

	return
}
