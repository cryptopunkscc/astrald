package frames

import (
	"encoding/binary"
	"fmt"
	"io"

	"github.com/cryptopunkscc/astrald/astral"
)

var _ Frame = &Migrate{}

type Migrate struct {
	Nonce astral.Nonce
}

func (frame *Migrate) ReadFrom(r io.Reader) (n int64, err error) {
	var opcode uint8

	err = binary.Read(r, binary.BigEndian, &opcode)
	if err != nil {
		return
	}
	n += 1

	if opcode != opMigrate {
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

func (frame *Migrate) WriteTo(w io.Writer) (n int64, err error) {
	err = binary.Write(w, binary.BigEndian, uint8(opMigrate))
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

func (frame *Migrate) String() string {
	return fmt.Sprintf("migrate(%s)", frame.Nonce)
}
