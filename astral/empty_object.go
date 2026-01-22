package astral

import "io"

type EmptyObject struct{}

// astral

func (EmptyObject) WriteTo(w io.Writer) (n int64, err error) {
	return 0, nil
}

func (*EmptyObject) ReadFrom(r io.Reader) (n int64, err error) {
	return 0, err
}

// json

func (a *EmptyObject) UnmarshalJSON(bytes []byte) error {
	return nil
}

func (a EmptyObject) MarshalJSON() ([]byte, error) {
	return jsonNull, nil
}

// text

func (a *EmptyObject) UnmarshalText(text []byte) error {
	return nil
}

func (a EmptyObject) MarshalText() (text []byte, err error) {
	return []byte{}, nil
}
