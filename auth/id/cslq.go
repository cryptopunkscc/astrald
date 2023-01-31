package id

import (
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/cryptopunkscc/astrald/cslq"
)

const cslqPattern = "[33]c"

func (id *Identity) UnmarshalCSLQ(dec *cslq.Decoder) error {
	var bytes []byte

	if err := dec.Decode(cslqPattern, &bytes); err != nil {
		return err
	}

	parsed, err := ParsePublicKey(bytes)
	if err != nil {
		return err
	}

	*id = parsed

	return nil
}

func (id Identity) MarshalCSLQ(enc *cslq.Encoder) error {
	var serialized []byte
	if id.IsZero() {
		serialized = make([]byte, btcec.PubKeyBytesLenCompressed)
	} else {
		serialized = id.PublicKey().SerializeCompressed()
	}
	return enc.Encode(cslqPattern, serialized)
}
