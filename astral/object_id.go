package astral

import (
	"bytes"
	"database/sql/driver"
	"encoding/base32"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
)

const idPrefix = "data1"
const zBase32CharSet = "ybndrfg8ejkmcpqxot1uwisza345h769"

var zBase32Encoding = base32.NewEncoding(zBase32CharSet)

type ObjectID struct {
	Size uint64
	Hash [32]byte
}

func ParseID(s string) (id *ObjectID, err error) {
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

	id = &ObjectID{}
	id.Size = ByteOrder.Uint64(data[0:8])
	copy(id.Hash[:], data[8:40])

	return
}

// astral

func (ObjectID) ObjectType() string {
	return "object_id.sha256"
}

func (id *ObjectID) WriteTo(w io.Writer) (n int64, err error) {
	if id.IsZero() {
		m, err := w.Write(make([]byte, 40))
		return int64(m), err
	}

	err = binary.Write(w, ByteOrder, id.Size)
	if err != nil {
		return
	}
	n += 8

	n2, err := w.Write(id.Hash[:])
	n += int64(n2)

	return
}

func (id *ObjectID) ReadFrom(r io.Reader) (n int64, err error) {
	err = binary.Read(r, ByteOrder, &id.Size)
	if err != nil {
		return
	}
	n += 8

	n2, err := io.ReadFull(r, id.Hash[:])
	n += int64(n2)

	return
}

// json

func (id ObjectID) MarshalJSON() ([]byte, error) {
	if id.IsZero() {
		return []byte("\"\""), nil
	}

	return []byte(fmt.Sprintf("\"%s\"", id.String())), nil
}

func (id *ObjectID) UnmarshalJSON(b []byte) error {
	var s string
	var jsonDec = json.NewDecoder(bytes.NewReader(b))

	var err = jsonDec.Decode(&s)
	if err != nil {
		return err
	}

	parsed, err := ParseID(s)
	if err != nil {
		return err
	}

	*id = *parsed

	return nil
}

// text

func (id ObjectID) MarshalText() (text []byte, err error) {
	return []byte(id.String()), nil
}

func (id *ObjectID) UnmarshalText(text []byte) (err error) {
	parsed, err := ParseID(string(text))
	if err != nil {
		return err
	}
	*id = *parsed
	return
}

// sql

func (id ObjectID) Value() (driver.Value, error) {
	return id.String(), nil
}

func (id *ObjectID) Scan(src any) error {
	if src == nil {
		*id = ObjectID{}
		return nil
	}

	str, ok := src.(string)
	if !ok {
		return errors.New("typecast failed")
	}

	parsed, err := ParseID(str)
	if err != nil {
		return err
	}

	*id = *parsed

	return nil
}

// ...

func (id ObjectID) String() string {
	var b [40]byte
	ByteOrder.PutUint64(b[0:8], id.Size)
	copy(b[8:], id.Hash[0:32])
	enc := zBase32Encoding.EncodeToString(b[:])
	enc = strings.TrimLeft(enc, zBase32CharSet[0:1])
	return idPrefix + enc
}

func (id *ObjectID) IsEqual(other *ObjectID) bool {
	if id.IsZero() {
		return other.IsZero()
	}

	if id.Size != other.Size {
		return false
	}

	return bytes.Compare(id.Hash[:], other.Hash[:]) == 0
}

func (id *ObjectID) IsZero() bool {
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

func init() {
	_ = DefaultBlueprints.Add(&ObjectID{})
}
