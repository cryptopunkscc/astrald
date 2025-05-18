package object

import (
	"bytes"
	"encoding/base32"
	"encoding/binary"
	"errors"
	"io"
	"strings"
)

const idPrefix = "data1"
const zBase32CharSet = "ybndrfg8ejkmcpqxot1uwisza345h769"

var zBase32Encoding = base32.NewEncoding(zBase32CharSet)

type ID struct {
	Size uint64
	Hash [32]byte
}

func ParseID(s string) (id *ID, err error) {
	// Check and trim the prefix
	if !strings.HasPrefix(s, idPrefix) {
		return nil, errors.New("invalid prefix")
	}
	s = strings.TrimPrefix(s, idPrefix)

	// Pad with missing leading zeros
	z := max(64-len(s), 0)
	padded := strings.Repeat(zBase32CharSet[0:1], z) + s

	var data [40]byte
	n, err := zBase32Encoding.Decode(data[:], []byte(padded))
	if err != nil {
		return nil, err
	}
	if n != 40 {
		return nil, errors.New("invalid data length")
	}

	id = &ID{}
	id.Size = binary.BigEndian.Uint64(data[0:8])
	copy(id.Hash[:], data[8:40])

	return
}

func (id *ID) WriteTo(w io.Writer) (n int64, err error) {
	if id.IsZero() {
		m, err := w.Write(make([]byte, 40))
		return int64(m), err
	}

	err = binary.Write(w, binary.BigEndian, id.Size)
	if err != nil {
		return
	}
	n += 8

	n2, err := w.Write(id.Hash[:])
	n += int64(n2)

	return
}

func (id *ID) ReadFrom(r io.Reader) (n int64, err error) {
	err = binary.Read(r, binary.BigEndian, &id.Size)
	if err != nil {
		return
	}
	n += 8

	n2, err := io.ReadFull(r, id.Hash[:])
	n += int64(n2)

	return
}

func (id ID) String() string {
	var b [40]byte
	binary.BigEndian.PutUint64(b[0:8], id.Size)
	copy(b[8:], id.Hash[0:32])
	enc := zBase32Encoding.EncodeToString(b[:])
	enc = strings.TrimLeft(enc, zBase32CharSet[0:1])
	return idPrefix + enc
}

func (id *ID) IsEqual(other *ID) bool {
	if id.IsZero() {
		return other.IsZero()
	}

	if id.Size != other.Size {
		return false
	}

	return bytes.Compare(id.Hash[:], other.Hash[:]) == 0
}

func (id *ID) IsZero() bool {
	if id == nil {
		return true
	}

	for _, b := range id.Hash {
		if b != 0 {
			return false
		}
	}
	return true
}

func (ID) ObjectType() string {
	return "astral.object_id.sha256"
}

func (id *ID) UnmarshalText(text []byte) (err error) {
	parsed, err := ParseID(string(text))
	if err != nil {
		return err
	}
	*id = *parsed
	return
}

func (id ID) MarshalText() (text []byte, err error) {
	return []byte(id.String()), nil
}
