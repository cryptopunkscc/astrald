package fid

import (
	"encoding/base32"
	"encoding/binary"
	"errors"
	"io"
	"strings"
)

const (
	SizeLen = 8
	HashLen = 32
	Size    = 40
)

const idPrefix = "id1"
const zBase32CharSet = "ybndrfg8ejkmcpqxot1uwisza345h769"

var zBase32Encoding = base32.NewEncoding(zBase32CharSet)

type ID struct {
	Size uint64
	Hash [32]byte
}

func (id ID) Pack() [Size]byte {
	var b [Size]byte
	binary.BigEndian.PutUint64(b[0:8], id.Size)
	copy(b[8:], id.Hash[0:32])
	return b
}

func Unpack(data [Size]byte) (id ID) {
	id.Size = binary.BigEndian.Uint64(data[0:8])
	copy(id.Hash[:], data[8:Size])
	return
}

func Read(reader io.Reader) (id ID, data [Size]byte, err error) {
	_, err = reader.Read(data[:])
	if err != nil {
		return
	}
	id = Unpack(data)
	return
}

func (id ID) String() string {
	packed := id.Pack()
	enc := zBase32Encoding.EncodeToString(packed[:])
	enc = strings.TrimLeft(enc, zBase32CharSet[0:1])
	return idPrefix + enc
}

func Parse(s string) (id ID, err error) {
	// Check and trim the prefix
	if !strings.HasPrefix(s, idPrefix) {
		return ID{}, errors.New("invalid prefix")
	}
	s = strings.TrimPrefix(s, idPrefix)

	// Pad with missing leading zeros
	z := 64 - len(s)
	padded := strings.Repeat(zBase32CharSet[0:1], z) + s

	var data [Size]byte
	n, err := zBase32Encoding.Decode(data[:], []byte(padded))
	if err != nil {
		return ID{}, err
	}
	if n != Size {
		return ID{}, errors.New("invalid data length")
	}

	return Unpack(data), nil
}
