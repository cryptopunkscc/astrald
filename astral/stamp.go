package astral

import (
	"encoding/binary"
	"errors"
	"io"
)

// Stamp is a special binary object always prepended to the canonical form of an object.
type Stamp struct{}

// magic is a const number at the very beginning of the object header
const magic = uint32(0x41444330)

func (*Stamp) ObjectType() string { return "stamp" }

func (s Stamp) WriteTo(w io.Writer) (n int64, err error) {
	err = binary.Write(w, binary.BigEndian, uint32(magic))
	if err == nil {
		n = 4
	}
	return
}

func (s *Stamp) ReadFrom(r io.Reader) (n int64, err error) {
	var m uint32
	err = binary.Read(r, binary.BigEndian, &m)
	if err == nil {
		n = 4
	}
	if m != magic {
		err = errors.New("invalid magic bytes")
		return
	}

	return
}

func init() {
	_ = DefaultBlueprints.Add(&Stamp{})
}
