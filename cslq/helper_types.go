package cslq

import (
	"time"
)

type StringC string
type StringS string
type StringL string
type StringQ string

type BufferC []byte
type BufferS []byte
type BufferL []byte
type BufferQ []byte

type Time time.Time

func (s StringC) FormatCSLQ() string {
	return "[c]c"
}

func (s StringS) FormatCSLQ() string {
	return "[s]c"
}

func (s StringL) FormatCSLQ() string {
	return "[l]c"
}

func (s StringQ) FormatCSLQ() string {
	return "[q]c"
}

func (s BufferC) FormatCSLQ() string {
	return "[c]c"
}

func (s BufferS) FormatCSLQ() string {
	return "[s]c"
}

func (s BufferL) FormatCSLQ() string {
	return "[l]c"
}

func (s BufferQ) FormatCSLQ() string {
	return "[q]c"
}

func (t *Time) UnmarshalCSLQ(dec *Decoder) (err error) {
	var v int64
	err = dec.Decodef("q", &v)
	*(*time.Time)(t) = time.Unix(0, v)
	return
}

func (t *Time) MarshalCSLQ(enc *Encoder) error {
	return enc.Encodef("q", (*time.Time)(t).UnixNano())
}

func (t Time) Time() time.Time {
	return time.Time(t)
}
