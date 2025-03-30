package astral

import "io"

type Ack struct{}

var _ Object = &Ack{}

func (Ack) ObjectType() string { return "ack" }

func (Ack) WriteTo(w io.Writer) (n int64, err error) {
	return 0, nil
}

func (*Ack) ReadFrom(r io.Reader) (n int64, err error) {
	return 0, err
}
