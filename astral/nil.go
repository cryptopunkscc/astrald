package astral

import "io"

// Nil is a pseudo-object that represents the nil value.
type Nil struct{}

var _ Object = &Nil{}

// astral

func (Nil) ObjectType() string { return "nil" }

func (Nil) WriteTo(w io.Writer) (n int64, err error) {
	return 0, nil
}

func (*Nil) ReadFrom(r io.Reader) (n int64, err error) {
	return 0, err
}

// json

func (n Nil) UnmarshalJSON(bytes []byte) error {
	return nil
}

func (n Nil) MarshalJSON() ([]byte, error) {
	return []byte("\"\""), nil
}

// text

func (n Nil) UnmarshalText(text []byte) error {
	return nil
}

func (n Nil) MarshalText() (text []byte, err error) {
	return []byte{}, nil
}

func init() {
	DefaultBlueprints.Add(&Nil{})
}
