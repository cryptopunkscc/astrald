package frames

import (
	"encoding/binary"
	"fmt"
	"github.com/cryptopunkscc/astrald/astral"
	"io"
)

var _ Frame = &Data{}

// Data is a Frame for transporting data
type Data struct {
	Nonce   astral.Nonce
	Payload []byte
}

func (frame *Data) ReadFrom(r io.Reader) (n int64, err error) {
	var opcode uint8

	err = binary.Read(r, binary.BigEndian, &opcode)
	if err != nil {
		return
	}
	n += 1

	if opcode != opData {
		err = ErrInvalidOpcode
		return
	}

	err = binary.Read(r, binary.BigEndian, &frame.Nonce)
	if err != nil {
		return
	}
	n += 8

	var plen uint16
	err = binary.Read(r, binary.BigEndian, &plen)
	if err != nil {
		return
	}
	n += 2

	frame.Payload = make([]byte, plen)
	var m int
	m, err = io.ReadFull(r, frame.Payload)
	n += int64(m)

	return
}

func (frame *Data) WriteTo(w io.Writer) (n int64, err error) {
	err = binary.Write(w, binary.BigEndian, uint8(opData))
	if err != nil {
		return
	}
	n += 1

	err = binary.Write(w, binary.BigEndian, frame.Nonce)
	if err != nil {
		return
	}
	n += 8

	err = binary.Write(w, binary.BigEndian, uint16(len(frame.Payload)))
	if err != nil {
		return
	}
	n += 2

	var m int
	m, err = w.Write(frame.Payload)
	n += int64(m)

	return
}

func (frame *Data) String() string {
	return fmt.Sprintf("data(%s,[%d])", frame.Nonce.String(), len(frame.Payload))
}
