package astral

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
)

type Nonce uint64

func NewNonce() (nonce Nonce) {
	binary.Read(rand.Reader, binary.BigEndian, &nonce)
	return
}

func (n Nonce) String() string {
	return fmt.Sprintf("%016x", uint64(n))
}
