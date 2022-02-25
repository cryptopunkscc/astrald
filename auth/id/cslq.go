package id

import "github.com/cryptopunkscc/astrald/cslq"

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
	return enc.Encode(cslqPattern, id.PublicKey().SerializeCompressed())
}
