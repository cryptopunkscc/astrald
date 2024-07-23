package frames

import (
	"encoding/binary"
	"fmt"
	"github.com/cryptopunkscc/astrald/astral"
	"io"
)

var _ Frame = &Read{}

// Read is a frame requesting more data for the nonce
type Read struct {
	Nonce astral.Nonce
	Len   uint32
}

func (frame *Read) ReadFrom(r io.Reader) (n int64, err error) {
	var opcode uint8

	err = binary.Read(r, binary.BigEndian, &opcode)
	if err != nil {
		return
	}
	n += 1

	if opcode != opRead {
		err = ErrInvalidOpcode
		return
	}

	err = binary.Read(r, binary.BigEndian, &frame.Nonce)
	if err != nil {
		return
	}
	n += 8

	err = binary.Read(r, binary.BigEndian, &frame.Len)
	if err != nil {
		return
	}
	n += 4

	return
}

func (frame *Read) WriteTo(w io.Writer) (n int64, err error) {
	err = binary.Write(w, binary.BigEndian, uint8(opRead))
	if err != nil {
		return
	}
	n += 1

	err = binary.Write(w, binary.BigEndian, frame.Nonce)
	if err != nil {
		return
	}
	n += 8

	err = binary.Write(w, binary.BigEndian, frame.Len)
	if err != nil {
		return
	}
	n += 4

	return
}

func (frame *Read) String() string {
	return fmt.Sprintf("read(%s,%d)", frame.Nonce.String(), frame.Len)
}
