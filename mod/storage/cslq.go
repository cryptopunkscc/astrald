package storage

import "github.com/cryptopunkscc/astrald/cslq"

func (o *PurgeOpts) UnmarshalCSLQ(*cslq.Decoder) (err error) {
	return nil
}

func (o *PurgeOpts) MarshalCSLQ(*cslq.Encoder) error {
	return nil
}

func (o *OpenOpts) UnmarshalCSLQ(dec *cslq.Decoder) (err error) {
	return dec.Decodef("q c c", &o.Offset, &o.Virtual, &o.Network)
}

func (o *OpenOpts) MarshalCSLQ(enc *cslq.Encoder) error {
	return enc.Encodef("q c c", o.Offset, o.Virtual, o.Network)
}

func (o *CreateOpts) UnmarshalCSLQ(dec *cslq.Decoder) (err error) {
	return dec.Decodef("l", &o.Alloc)
}

func (o *CreateOpts) MarshalCSLQ(enc *cslq.Encoder) error {
	return enc.Encodef("l", o.Alloc)
}

func (o *ReaderInfo) UnmarshalCSLQ(dec *cslq.Decoder) (err error) {
	return dec.Decodef("[c]c", &o.Name)
}

func (o *ReaderInfo) MarshalCSLQ(enc *cslq.Encoder) error {
	return enc.Encodef("[c]c", o.Name)
}
