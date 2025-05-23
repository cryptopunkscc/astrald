package astral

import (
	"crypto/rand"
	"database/sql/driver"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"strconv"
)

type Nonce uint64

func NewNonce() (nonce Nonce) {
	binary.Read(rand.Reader, binary.BigEndian, &nonce)
	return
}

// astral

func (Nonce) ObjectType() string { return "nonce64" }

func (nonce Nonce) WriteTo(w io.Writer) (n int64, err error) {
	err = binary.Write(w, binary.BigEndian, uint64(nonce))
	if err == nil {
		n = 8
	}
	return
}

func (nonce *Nonce) ReadFrom(r io.Reader) (n int64, err error) {
	err = binary.Read(r, binary.BigEndian, nonce)
	if err == nil {
		n = 8
	}
	return
}

// text

func (nonce Nonce) MarshalText() (text []byte, err error) {
	return []byte(nonce.String()), nil
}

func (nonce *Nonce) UnmarshalText(text []byte) (err error) {
	u, err := strconv.ParseUint(string(text), 16, 64)
	*nonce = Nonce(u)
	return
}

// sql

func (nonce Nonce) Value() (driver.Value, error) {
	return fmt.Sprintf("%016x", uint64(nonce)), nil
}

func (nonce *Nonce) Scan(src any) error {
	v, ok := src.(string)
	if !ok {
		return errors.New("typcast failed")
	}

	u, err := strconv.ParseUint(v, 16, 64)
	if err != nil {
		return err
	}

	*nonce = Nonce(u)
	return nil
}

// ...

func (nonce Nonce) String() string {
	return fmt.Sprintf("%016x", uint64(nonce))
}

func init() {
	var n Nonce
	_ = DefaultBlueprints.Add(&n)
}
