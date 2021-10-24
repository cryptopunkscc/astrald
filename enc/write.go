package enc

import (
	"encoding/binary"
	"errors"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/infra"
	"io"
)

func Write(w io.Writer, i interface{}) error {
	return binary.Write(w, binary.BigEndian, i)
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

	_, err = w.Write([]byte{byte(l)})
	if err != nil {
		return err
	}

	_, err = w.Write(bytes)
	return err
}

func WriteIdentity(w io.Writer, id id.Identity) error {
	_, err := w.Write(id.PublicKey().SerializeCompressed())
	return err
}

func WriteAddr(w io.Writer, addr infra.Addr) error {
	switch addr.Network() {
	case "inet":
		if err := Write(w, uint8(0)); err != nil {
			return err
		}
	case "tor":
		if err := Write(w, uint8(1)); err != nil {
			return err
		}
	default:
		if err := Write(w, uint8(255)); err != nil {
			return err
		}
		if err := WriteL8String(w, addr.Network()); err != nil {
			return err
		}
	}
	if err := WriteL8Bytes(w, addr.Pack()); err != nil {
		return err
	}
	return nil
}
