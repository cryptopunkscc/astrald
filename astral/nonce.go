package astral

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"io"
	"strconv"
)

type Nonce uint64

func NewNonce() (nonce Nonce) {
	binary.Read(rand.Reader, binary.BigEndian, &nonce)
	return
}

func (Nonce) ObjectType() string { return "astral.nonce64" }

func (nonce Nonce) WriteTo(w io.Writer) (n int64, err error) {
	err = binary.Write(w, binary.BigEndian, uint64(nonce))
	if err == nil {
		n = 8
	}
	return
}

func (nonce *Nonce) ReadFrom(r io.Reader) (n int64, err error) {
	err = binary.Read(r, binary.BigEndian, &nonce)
	if err == nil {
		n = 8
	}
	return
}

func (nonce Nonce) String() string {
	return fmt.Sprintf("%016x", uint64(nonce))
}

func (nonce Nonce) MarshalText() (text []byte, err error) {
	return []byte(nonce.String()), nil
}

func (nonce *Nonce) UnmarshalText(text []byte) (err error) {
	u, err := strconv.ParseUint(string(text), 16, 64)
	*nonce = Nonce(u)
	return
}

func init() {
	var n Nonce
	DefaultBlueprints.Add(&n)
}
