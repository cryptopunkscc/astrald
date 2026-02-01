package crypto

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/crypto"
)

type ObjectSigner struct {
	mod    *Module
	scheme string
	key    *crypto.PublicKey
}

var _ crypto.ObjectSigner = &ObjectSigner{}

func (s *ObjectSigner) SignObject(ctx *astral.Context, object crypto.SignableObject) (*crypto.Signature, error) {
	signer, err := s.mod.HashSigner(s.key, s.scheme)
	if err != nil {
		return nil, err
	}

	return signer.SignHash(ctx, object.SignableHash())
}

type TextObjectSigner struct {
	mod    *Module
	scheme string
	key    *crypto.PublicKey
}

var _ crypto.TextObjectSigner = &TextObjectSigner{}

func (s *TextObjectSigner) SignTextObject(ctx *astral.Context, object crypto.SignableTextObject) (*crypto.Signature, error) {
	signer, err := s.mod.TextSigner(s.key, s.scheme)
	if err != nil {
		return nil, err
	}

	return signer.SignText(ctx, s.mod.formatSignableText(object))
}
