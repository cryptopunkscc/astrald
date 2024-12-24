package ether

import (
	"crypto/ecdsa"
	"errors"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/streams"
	"io"
)

var _ astral.Object = &SignedBroadcast{}

type SignedBroadcast struct {
	Broadcast
	Signature astral.Bytes16
}

func (SignedBroadcast) ObjectType() string {
	return "astrald.mod.ether.signed_broadcast"
}

func (b SignedBroadcast) WriteTo(w io.Writer) (n int64, err error) {
	return streams.WriteAllTo(w,
		b.Broadcast,
		b.Signature,
	)
}

func (b *SignedBroadcast) ReadFrom(r io.Reader) (n int64, err error) {
	return streams.ReadAllFrom(r,
		&b.Broadcast,
		&b.Signature,
	)
}

func (b SignedBroadcast) Hash() []byte {
	objectID, err := astral.ResolveObjectID(&b.Broadcast)
	if err != nil {
		return nil
	}
	return objectID.Hash[:]
}

func (b SignedBroadcast) VerifySig() error {
	switch {
	case b.Source.IsZero():
		return errors.New("source identity missing")
	case b.Signature == nil:
		return errors.New("signature missing")
	case !ecdsa.VerifyASN1(
		b.Source.PublicKey().ToECDSA(),
		b.Hash(),
		b.Signature,
	):
		return errors.New("signature is invalid")
	}
	return nil
}
