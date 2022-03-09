package cslq

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"io"
)

type OpByte uint8

func (op OpByte) Encode(w io.Writer, _ *Fifo) error {
	return binary.Write(w, byteOrder, uint8(op))
}

func (op OpByte) Decode(r io.Reader, _ *Fifo) error {
	var b uint8
	if err := binary.Read(r, byteOrder, &b); err != nil {
		return err
	}
	if uint8(op) != b {
		return errors.New("byte literal mismatch")
	}
	return nil
}

func (op OpByte) String() string {
	return "x" + hex.EncodeToString([]byte{byte(op)})
}
