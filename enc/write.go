package enc

import (
	"encoding/binary"
	"errors"
	"github.com/cryptopunkscc/astrald/auth/id"
	"io"
)

var encoding = binary.BigEndian

func Write(w io.Writer, i interface{}) error {
	return binary.Write(w, encoding, i)
}

func WriteL8String(w io.Writer, str string) error {
	return WriteL8Bytes(w, []byte(str))
}

func WriteL8Bytes(w io.Writer, bytes []byte) error {
	var err error
	var l = len(bytes)

	if l > 255 {
		return errors.New("data too long")
	}

	err = Write(w, uint8(l))
	if err != nil {
		return err
	}

	_, err = w.Write(bytes)
	return err
}

func WriteIdentity(w io.Writer, id id.Identity) error {
	// Zero identity is encoded as 33 zero bytes
	if id.IsZero() {
		var empty [33]byte
		_, err := w.Write(empty[:])
		return err
	}
	_, err := w.Write(id.PublicKey().SerializeCompressed())
	return err
}
