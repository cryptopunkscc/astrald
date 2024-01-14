package data

import "github.com/cryptopunkscc/astrald/cslq"

// Astral Data Container is a generic header for data. It contains a single string field
// describing the data type.

var adc0Format = "<x41x44x43x30>[c]c"

type ADC0Header string

func (t *ADC0Header) UnmarshalCSLQ(dec *cslq.Decoder) error {
	var err error
	var s string
	err = dec.Decodef(adc0Format, &s)
	*t = ADC0Header(s)
	return err
}

func (t ADC0Header) MarshalCSLQ(enc *cslq.Encoder) error {
	return enc.Encodef(adc0Format, t)
}
