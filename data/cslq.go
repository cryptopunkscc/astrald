package data

import "github.com/cryptopunkscc/astrald/cslq"

var _ cslq.Marshaler = &ID{}
var _ cslq.Unmarshaler = &ID{}

func (id *ID) UnmarshalCSLQ(dec *cslq.Decoder) (err error) {
	var buf [40]byte
	if err = dec.Decodef("[40]c", &buf); err != nil {
		return
	}

	*id = Unpack(buf)

	return
}

func (id ID) MarshalCSLQ(enc *cslq.Encoder) error {
	return enc.Encodef("[40]c", id.Pack())
}
