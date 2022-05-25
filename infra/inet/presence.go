package inet

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/cslq"
)

type presence struct {
	Identity id.Identity
	Port     int
	Flags    uint8
}

const (
	flagNone     = 0x00
	flagDiscover = 0x01
	flagBye      = 0x02
)

const presenceCSLQ = "x61 x70 x00 x00 v s c"

func (p presence) MarshalCSLQ(enc *cslq.Encoder) error {
	return enc.Encode(presenceCSLQ, p.Identity, p.Port, p.Flags)
}

func (p *presence) UnmarshalCSLQ(dec *cslq.Decoder) error {
	return dec.Decode(presenceCSLQ, &p.Identity, &p.Port, &p.Flags)
}
