package crypto

import (
	"errors"

	"github.com/cryptopunkscc/astrald/astral"
)

// NilEngine returns ErrUnsupported for all operations. Embed it in your engine to
// avoid having to explicitly implement unsupported interface methods:
//
//	type MyEngine struct{
//	  NilEngine
//	}
//
//	var _ Engine = &MyEngine{} // no error
//
//	func (MyEngine) PublicKey(*PrivateKey) (*PublicKey, error) {
//	  // ...
//	}
type NilEngine struct{}

func (NilEngine) PublicKey(*astral.Context, *PrivateKey) (*PublicKey, error) {
	return nil, errors.ErrUnsupported
}

func (NilEngine) HashSigner(*PublicKey, string) (HashSigner, error) {
	return nil, errors.ErrUnsupported
}

func (NilEngine) VerifyHashSignature(*PublicKey, *Signature, []byte) error {
	return errors.ErrUnsupported
}

func (NilEngine) TextSigner(*PublicKey, string) (TextSigner, error) {
	return nil, errors.ErrUnsupported
}

func (NilEngine) VerifyTextSignature(*PublicKey, *Signature, string) error {
	return errors.ErrUnsupported
}
