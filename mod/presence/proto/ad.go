package proto

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/cslq"
)

const adFormat = "x61 x70 x00 x00 v [c]c s c"

const (
	FlagDiscover = 1 << uint8(iota)
)

type Ad struct {
	Identity id.Identity
	Alias    string
	Port     int
	Flags    uint8
}

func (ad *Ad) MarshalCSLQ(enc *cslq.Encoder) error {
	return enc.Encodef(adFormat, ad.Identity, ad.Alias, ad.Port, ad.Flags)
}

func (ad *Ad) UnmarshalCSLQ(dec *cslq.Decoder) error {
	return dec.Decodef(adFormat, &ad.Identity, &ad.Alias, &ad.Port, &ad.Flags)
}
