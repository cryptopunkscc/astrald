package adc

import (
	"errors"
	"github.com/cryptopunkscc/astrald/cslq"
	"io"
)

// ADC - Astral Data Container is a minimal, generic data header used for astral data objects.
// It consists of fixed bytes "ADC0" and a length-encoded (8 bits) string describing the data type.

var format = "<x41x44x43x30>[c]c"

type Header string

func (header *Header) UnmarshalCSLQ(dec *cslq.Decoder) error {
	var err error
	var s string
	err = dec.Decodef(format, &s)
	*header = Header(s)
	return err
}

func (header Header) MarshalCSLQ(enc *cslq.Encoder) error {
	if len(header) == 0 {
		return nil
	}

	return enc.Encodef(format, header)
}

func (header Header) String() string { return string(header) }

func ReadHeader(reader io.Reader) (Header, error) {
	var header Header
	var err = cslq.Decode(reader, "v", &header)
	return header, err
}

func WriteHeader(writer io.Writer, header Header) error {
	return cslq.Encode(writer, "v", header)
}

func ExpectHeader(reader io.Reader, header Header) error {
	h, err := ReadHeader(reader)
	if err != nil {
		return err
	}
	if h != header {
		return errors.New("header mismatch")
	}
	return nil
}
