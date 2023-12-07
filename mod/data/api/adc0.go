package data

import "github.com/cryptopunkscc/astrald/cslq"

var adc0Format = "<x41x44x43x30>[c]c"

type ADC0Header struct {
	Type string
}

func (h *ADC0Header) UnmarshalCSLQ(dec *cslq.Decoder) error {
	return dec.Decodef(adc0Format, &h.Type)
}

func (h *ADC0Header) MarshalCSLQ(enc *cslq.Encoder) error {
	return enc.Encodef(adc0Format, h.Type)
}
