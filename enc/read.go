package enc

import (
	"encoding/binary"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/infra"
	"io"
)

func ReadUint8(r io.Reader) (i uint8, err error) {
	err = binary.Read(r, binary.BigEndian, &i)
	return
}

func ReadUint16(r io.Reader) (i uint16, err error) {
	err = binary.Read(r, binary.BigEndian, &i)
	return
}

func ReadL8String(r io.Reader) (string, error) {
	bytes, err := ReadL8Bytes(r)
	return string(bytes), err
}

func ReadL8Bytes(r io.Reader) ([]byte, error) {
	l, err := ReadUint8(r)
	if err != nil {
		return nil, err
	}

	buf := make([]byte, l)
	_, err = io.ReadFull(r, buf)
	if err != nil {
		return nil, err
	}

	return buf, nil
}

func ReadIdentity(r io.Reader) (id.Identity, error) {
	buf := make([]byte, 33)
	_, err := io.ReadFull(r, buf)
	if err != nil {
		return id.Identity{}, err
	}
	return id.ParsePublicKey(buf)
}

func ReadAddr(r io.Reader) (infra.Addr, error) {
	net, err := ReadUint8(r)
	if err != nil {
		return nil, err
	}

	var netName string

	switch net {
	case 0:
		netName = "inet"
	case 1:
		netName = "tor"
	case 255:
		netName, err = ReadL8String(r)
		if err != nil {
			return nil, err
		}
	}

	data, err := ReadL8Bytes(r)
	if err != nil {
		return nil, err
	}

	return astral.Unpack(netName, data)
}
