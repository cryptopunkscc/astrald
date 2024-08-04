package keys

import (
	"bytes"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/cslq"
	"io"
)

const KeyTypeIdentity = "ecdsa-secp256k1"

var _ astral.Object = &PrivateKey{}

type PrivateKey struct {
	Type  string
	Bytes []byte
}

func (PrivateKey) ObjectType() string {
	return "keys.private_key"
}

func (p PrivateKey) WriteTo(w io.Writer) (n int64, err error) {
	var b = &bytes.Buffer{}
	cslq.Encode(b, "[c]c[c]c", p.Type, p.Bytes)
	m, err := w.Write(b.Bytes())
	return int64(m), err
}

func (p *PrivateKey) ReadFrom(r io.Reader) (n int64, err error) {
	err = cslq.Decode(r, "[c]c[c]c", &p.Type, &p.Bytes)
	return
}
