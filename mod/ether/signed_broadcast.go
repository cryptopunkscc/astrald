package ether

import (
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
