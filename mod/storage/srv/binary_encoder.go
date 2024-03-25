package proto

import (
	"encoding/binary"
	"io"
)

type BinaryEncoderCloser struct {
	BinaryEncoder
	io.Closer
}

func NewBinaryEncoderCloser(rwc io.ReadWriteCloser) BinaryEncoderCloser {
	return BinaryEncoderCloser{BinaryEncoder: NewBinaryEncoder(rwc), Closer: rwc}
}

type BinaryEncoder struct {
	io.ReadWriter
}

func NewBinaryEncoder(rw io.ReadWriter) BinaryEncoder {
	return BinaryEncoder{rw}
}

func (enc BinaryEncoder) Encode(args ...any) (err error) {
	for _, arg := range args {
		if err = binary.Write(enc, binary.BigEndian, arg); err != nil {
			return
		}
	}
	return
}

func (enc BinaryEncoder) Decode(args ...any) (err error) {
	for _, arg := range args {
		if err = binary.Read(enc, binary.BigEndian, arg); err != nil {
			return
		}
	}
	return
}
