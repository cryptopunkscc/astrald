package astral

import "io"

type Ack struct{}

var _ Object = &Ack{}

// astral

func (Ack) ObjectType() string { return "ack" }

func (Ack) WriteTo(w io.Writer) (n int64, err error) {
	return 0, nil
}

func (*Ack) ReadFrom(r io.Reader) (n int64, err error) {
	return 0, err
}

// json

func (a Ack) UnmarshalJSON(bytes []byte) error {
	return nil
}

func (a Ack) MarshalJSON() ([]byte, error) {
	return []byte("\"\""), nil
}

// text

func (a Ack) UnmarshalText(text []byte) error {
	return nil
}

func (a Ack) MarshalText() (text []byte, err error) {
	return []byte{}, nil
}

func init() {
	DefaultBlueprints.Add(&Ack{})
}
