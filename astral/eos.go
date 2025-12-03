package astral

import "io"

var _ Object = &EOS{}

type EOS struct {
}

func (E EOS) ObjectType() string {
	return "eos"
}

func (EOS) WriteTo(w io.Writer) (n int64, err error) {
	return 0, nil
}

func (*EOS) ReadFrom(r io.Reader) (n int64, err error) {
	return 0, err
}

// json

func (a EOS) UnmarshalJSON(bytes []byte) error {
	return nil
}

func (a EOS) MarshalJSON() ([]byte, error) {
	return []byte("\"\""), nil
}

// text

func (a EOS) UnmarshalText(text []byte) error {
	return nil
}

func (a EOS) MarshalText() (text []byte, err error) {
	return []byte{}, nil
}

func init() {
	DefaultBlueprints.Add(&EOS{})
}
