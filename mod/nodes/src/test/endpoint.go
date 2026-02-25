package nodes

import (
	"bytes"
	"io"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/exonet"
)

type Endpoint struct {
	Addr string
}

var _ exonet.Endpoint = &Endpoint{}

func (t Endpoint) ObjectType() string {
	return "test.endpoint"
}

func (t Endpoint) WriteTo(w io.Writer) (n int64, err error) {
	return astral.Struct(t).WriteTo(w)
}

func (t *Endpoint) ReadFrom(r io.Reader) (n int64, err error) {
	return astral.Struct(t).ReadFrom(r)
}

func (t Endpoint) Network() string {
	return "test"
}

func (t Endpoint) Address() string {
	return t.Addr
}

func (t Endpoint) Pack() []byte {
	var b = &bytes.Buffer{}
	if _, err := t.WriteTo(b); err != nil {
		return nil
	}
	return b.Bytes()
}
