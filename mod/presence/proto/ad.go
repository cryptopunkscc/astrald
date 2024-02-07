package proto

import (
	"crypto/sha256"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/cslq"
)

const adFormat = "x61 x70 x00 x00 v [c]c s [c][c]c [c]c"
const hashFormat = "x61 x70 x00 x00 v [c]c s [c][c]c"

const (
	FlagDiscover = 1 << uint8(iota)
)

type Ad struct {
	Identity id.Identity
	Alias    string
	Port     int
	Flags    []string
	Sig      []byte
}

func (ad *Ad) MarshalCSLQ(enc *cslq.Encoder) error {
	return enc.Encodef(adFormat, ad.Identity, ad.Alias, ad.Port, ad.Flags, ad.Sig)
}

func (ad *Ad) UnmarshalCSLQ(dec *cslq.Decoder) error {
	return dec.Decodef(adFormat, &ad.Identity, &ad.Alias, &ad.Port, &ad.Flags, &ad.Sig)
}

func (ad *Ad) Hash() []byte {
	hash := sha256.New()

	err := cslq.Encode(hash, hashFormat, ad.Identity, ad.Alias, ad.Port, ad.Flags)
	if err != nil {
		return nil
	}

	return hash.Sum(nil)
}
