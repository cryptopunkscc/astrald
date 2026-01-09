package main

import (
	"bytes"
	"fmt"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/keys"
	"github.com/cryptopunkscc/astrald/resources"
)

// loadNodeIdentity loads node's identity from resources. Generates a new identity if we don't have one yet.
func loadNodeIdentity(resources resources.Resources) (*astral.Identity, error) {
	data, err := resources.Read(resNodeIdentity)

	// generate new identity if needed
	if err != nil {
		nodeID := astral.GenerateIdentity()

		var buf = &bytes.Buffer{}

		// encode the private key in canonical format
		_, err = astral.Encode(
			buf,
			(*astral.PrivateIdentity)(nodeID),
			astral.WithEncoder(astral.CanonicalTypeEncoder),
		)
		if err != nil {
			return nil, err
		}

		return nodeID, resources.Write(resNodeIdentity, buf.Bytes())
	}

	object, _, err := astral.Decode(bytes.NewReader(data), astral.Canonical())

	switch object := object.(type) {
	case *astral.PrivateIdentity:
		return (*astral.Identity)(object), nil

	case *keys.PrivateKey:
		return astral.IdentityFromPrivKeyBytes(object.Bytes)

	default:
		return nil, fmt.Errorf("unknown identity object: %s", object.ObjectType())
	}
}
