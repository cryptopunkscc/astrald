package cslq

import "bytes"

// Unmarshaler is an interface implemented by objects that can decode CSLQ representation of themselves.
type Unmarshaler interface {
	UnmarshalCSLQ(dec *Decoder) error
}

// Marshaler is an interface implemented by objects that can encode CSLQ representation of themselves.
type Marshaler interface {
	MarshalCSLQ(enc *Encoder) error
}

func Marshal(v any) ([]byte, error) {
	var buf = &bytes.Buffer{}
	var err = Encode(buf, "v", v)

	return buf.Bytes(), err
}

func Unmarshal(data []byte, v any) error {
	return Decode(bytes.NewReader(data), "v", v)
}
