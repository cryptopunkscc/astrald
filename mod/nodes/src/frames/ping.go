package frames

import (
	"encoding/binary"
	"fmt"
	"github.com/cryptopunkscc/astrald/astral"
	"io"
)

var _ Frame = &Ping{}

type Ping struct {
	Nonce astral.Nonce
	Pong  bool
}

func (frame *Ping) ReadFrom(r io.Reader) (n int64, err error) {
	var opcode uint8

	err = binary.Read(r, binary.BigEndian, &opcode)
	if err != nil {
		return
	}
	n += 1

	if opcode != opPing {
		err = ErrInvalidOpcode
		return
	}

	err = binary.Read(r, binary.BigEndian, &frame.Nonce)
	if err != nil {
		return
	}
	n += 8

	err = binary.Read(r, binary.BigEndian, &frame.Pong)
	if err != nil {
		return
	}
	n += 1

	return
}

func (frame *Ping) WriteTo(w io.Writer) (n int64, err error) {
	err = binary.Write(w, binary.BigEndian, uint8(opPing))
	if err != nil {
		return
	}
	n += 1

	err = binary.Write(w, binary.BigEndian, frame.Nonce)
	if err != nil {
		return
	}
	n += 8

	err = binary.Write(w, binary.BigEndian, frame.Pong)
	if err != nil {
		return
	}
	n += 1

	return
}

func (frame *Ping) String() string {
	if frame.Pong {
		return fmt.Sprintf("pong(%s)", frame.Nonce)
	}
	return fmt.Sprintf("ping(%s)", frame.Nonce)
}
