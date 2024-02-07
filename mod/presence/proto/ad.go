package proto

import (
	"crypto/sha256"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/cslq"
	"time"
)

const (
	hashFormat = "x61 x70 x00 x00 v [c]c s q [c][c]c"
	adFormat   = "x61 x70 x00 x00 v [c]c s q [c][c]c [c]c"
)

type Ad struct {
	Identity  id.Identity
	Alias     string
	Port      int
	ExpiresAt time.Time
	Flags     []string
	Sig       []byte
}

func (ad *Ad) MarshalCSLQ(enc *cslq.Encoder) error {
	return enc.Encodef(adFormat,
		ad.Identity,
		ad.Alias,
		ad.Port,
		ad.ExpiresAt.UnixNano(),
		ad.Flags,
		ad.Sig,
	)
}

func (ad *Ad) UnmarshalCSLQ(dec *cslq.Decoder) error {
	var t int64
	var err = dec.Decodef(adFormat,
		&ad.Identity,
		&ad.Alias,
		&ad.Port,
		&t,
		&ad.Flags,
		&ad.Sig,
	)
	ad.ExpiresAt = time.Unix(0, t)
	return err
}

func (ad *Ad) Hash() []byte {
	hash := sha256.New()

	err := cslq.Encode(hash, hashFormat,
		ad.Identity,
		ad.Alias,
		ad.Port,
		ad.ExpiresAt.UnixNano(),
		ad.Flags,
	)
	if err != nil {
		return nil
	}

	return hash.Sum(nil)
}
