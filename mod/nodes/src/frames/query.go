package frames

import (
	"encoding/binary"
	"fmt"
	"github.com/cryptopunkscc/astrald/astral"
	"io"
)

var _ Frame = &Query{}

type Query struct {
	Nonce  astral.Nonce
	Buffer uint32
	Query  string
}

func (frame *Query) ReadFrom(r io.Reader) (n int64, err error) {
	var opcode uint8

	err = binary.Read(r, binary.BigEndian, &opcode)
	if err != nil {
		return
	}
	n += 1

	if opcode != opQuery {
		err = ErrInvalidOpcode
		return
	}

	err = binary.Read(r, binary.BigEndian, &frame.Nonce)
	if err != nil {
		return
	}
	n += 8

	err = binary.Read(r, binary.BigEndian, &frame.Buffer)
	if err != nil {
		return
	}
	n += 4

	var plen uint16

	err = binary.Read(r, binary.BigEndian, &plen)
	if err != nil {
		return
	}
	n += 2

	var b = make([]byte, plen)
	var m int
	m, err = io.ReadFull(r, b)
	n += int64(m)
	frame.Query = string(b[:m])

	return
}

func (frame *Query) WriteTo(w io.Writer) (n int64, err error) {
	err = binary.Write(w, binary.BigEndian, uint8(opQuery))
	if err != nil {
		return
	}
	n += 1

	err = binary.Write(w, binary.BigEndian, frame.Nonce)
	if err != nil {
		return
	}
	n += 8

	err = binary.Write(w, binary.BigEndian, frame.Buffer)
	if err != nil {
		return
	}
	n += 4

	var plen = uint16(len(frame.Query))
	err = binary.Write(w, binary.BigEndian, plen)
	if err != nil {
		return
	}
	n += 2

	var m int
	m, err = w.Write([]byte(frame.Query))
	n += int64(m)

	return
}

func (frame *Query) String() string {
	return fmt.Sprintf("query(%s, '%s')", frame.Nonce.String(), frame.Query)
}
