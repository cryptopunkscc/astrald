package astral

import (
	"encoding/binary"
	"errors"
	"github.com/cryptopunkscc/astrald/object"
	"io"
)

// Object defines the basic interface of an astral object. An object must have a unique type and must be able to
// write/read its payload (the type is outside the payload) to/from a stream.
type Object interface {
	ObjectType() string
	io.WriterTo
	io.ReaderFrom
}

type ObjectReader interface {
	ReadObject(io.Reader) (Object, error)
}

// magic is a const number at the very beginning of the object header
const magic = uint32(0x41444330)

// ObjectHeader is the standard object header cotaining the object type. ObjectHeader is an Object itself.
type ObjectHeader string

func (ObjectHeader) ObjectType() string { return "astral.object_header" }

func (h ObjectHeader) WriteTo(w io.Writer) (n int64, err error) {
	err = binary.Write(w, binary.BigEndian, magic)
	if err != nil {
		return
	}
	n += 4

	err = binary.Write(w, binary.BigEndian, uint8(len(h)))
	if err != nil {
		return
	}
	n += 1

	n2, err := w.Write([]byte(h))
	n += int64(n2)
	return
}

func (h *ObjectHeader) ReadFrom(r io.Reader) (n int64, err error) {
	var m uint32
	err = binary.Read(r, binary.BigEndian, &m)
	if err != nil {
		return
	}
	n += 4

	if m != magic {
		err = errors.New("invalid magic bytes")
		return
	}

	var l uint8
	err = binary.Read(r, binary.BigEndian, &l)
	if err != nil {
		return
	}
	n += 1

	var buf = make([]byte, l)
	n2, err := io.ReadFull(r, buf)
	n += int64(n2)
	if err != nil {
		return
	}

	*h = ObjectHeader(buf)

	return
}

func (h ObjectHeader) String() string { return string(h) }

func DecodeObject(r io.Reader, o Object) (err error) {
	var head ObjectHeader
	_, err = head.ReadFrom(r)
	if err != nil {
		return
	}
	if head.String() != o.ObjectType() {
		return errors.New("object type mismatch")
	}
	_, err = o.ReadFrom(r)
	return err
}

func EncodeObject(w io.Writer, o Object) (err error) {
	_, err = ObjectHeader(o.ObjectType()).WriteTo(w)
	if err != nil {
		return
	}
	_, err = o.WriteTo(w)
	return
}

func ResolveObjectID(obj Object) (objectID object.ID, err error) {
	w := object.NewWriteResolver(nil)
	_, err = ObjectHeader(obj.ObjectType()).WriteTo(w)
	if err != nil {
		return
	}

	_, err = obj.WriteTo(w)
	if err != nil {
		return
	}

	return w.Resolve(), nil
}
