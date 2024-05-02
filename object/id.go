package object

import (
	"bytes"
	"encoding/base32"
	"encoding/binary"
	"errors"
	"strings"
)

const idPrefix = "data1"
const zBase32CharSet = "ybndrfg8ejkmcpqxot1uwisza345h769"

var zBase32Encoding = base32.NewEncoding(zBase32CharSet)

type ID struct {
	Size uint64
	Hash [32]byte
}

func (id ID) Pack() [40]byte {
	var b [40]byte
	binary.BigEndian.PutUint64(b[0:8], id.Size)
	copy(b[8:], id.Hash[0:32])
	return b
}

func (id ID) String() string {
	packed := id.Pack()
	enc := zBase32Encoding.EncodeToString(packed[:])
	enc = strings.TrimLeft(enc, zBase32CharSet[0:1])
	return idPrefix + enc
}

func (id ID) IsEqual(other ID) bool {
	if id.Size != other.Size {
		return false
	}

	if bytes.Compare(id.Hash[:], other.Hash[:]) != 0 {
		return false
	}

	return true
}

func (id ID) IsZero() bool {
	for _, b := range id.Hash {
		if b != 0 {
			return false
		}
	}
	return true
}

func Unpack(data [40]byte) (id ID) {
	id.Size = binary.BigEndian.Uint64(data[0:8])
	copy(id.Hash[:], data[8:40])
	return
}

func ParseID(s string) (id ID, err error) {
	// Check and trim the prefix
	if !strings.HasPrefix(s, idPrefix) {
		return ID{}, errors.New("invalid prefix")
	}
	s = strings.TrimPrefix(s, idPrefix)

	// Pad with missing leading zeros
	z := 64 - len(s)
	padded := strings.Repeat(zBase32CharSet[0:1], z) + s

	var data [40]byte
	n, err := zBase32Encoding.Decode(data[:], []byte(padded))
	if err != nil {
		return ID{}, err
	}
	if n != 40 {
		return ID{}, errors.New("invalid data length")
	}

	return Unpack(data), nil
}
