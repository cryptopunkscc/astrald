package astral

import (
	"crypto/rand"
	"database/sql/driver"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"
)

type Nonce uint64

func NewNonce() (nonce Nonce) {
	binary.Read(rand.Reader, ByteOrder, &nonce)
	return
}

func (Nonce) ObjectType() string { return "nonce64" }

// binary

func (nonce Nonce) WriteTo(w io.Writer) (n int64, err error) {
	err = binary.Write(w, ByteOrder, uint64(nonce))
	if err == nil {
		n = 8
	}
	return
}

func (nonce *Nonce) ReadFrom(r io.Reader) (n int64, err error) {
	err = binary.Read(r, ByteOrder, nonce)
	if err == nil {
		n = 8
	}
	return
}

// json

func (nonce Nonce) MarshalJSON() ([]byte, error) {
	return json.Marshal(strconv.FormatUint(uint64(nonce), 16))
}

func (nonce *Nonce) UnmarshalJSON(bytes []byte) (err error) {
	var s string
	var u uint64
	err = json.Unmarshal(bytes, &s)
	if err != nil {
		return
	}

	u, err = strconv.ParseUint(s, 16, 64)
	if err != nil {
		return
	}

	*nonce = Nonce(u)
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
	_ = Add(&n)
}
