package frames

import (
	"encoding/binary"
	"fmt"
	"github.com/cryptopunkscc/astrald/astral"
	"io"
)

var _ Frame = &Reset{}

type Reset struct {
	Nonce astral.Nonce
}

func (frame *Reset) ReadFrom(r io.Reader) (n int64, err error) {
	var opcode uint8

	err = binary.Read(r, binary.BigEndian, &opcode)
	if err != nil {
		return
	}
	n += 1

	if opcode != opReset {
		err = ErrInvalidOpcode
		return
	}

	err = binary.Read(r, binary.BigEndian, &frame.Nonce)
	if err != nil {
		return
	}
	n += 8

	return
}

func (frame *Reset) WriteTo(w io.Writer) (n int64, err error) {
	err = binary.Write(w, binary.BigEndian, uint8(opReset))
	if err != nil {
		return
	}
	n += 1

	err = binary.Write(w, binary.BigEndian, frame.Nonce)
	if err != nil {
		return
	}
	n += 8

	return
}

func (frame *Reset) String() string {
	return fmt.Sprintf("reset(%s)", frame.Nonce)
}
