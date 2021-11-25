package graph

import (
	"bytes"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/enc"
	"github.com/cryptopunkscc/astrald/infra"
	"github.com/jxskiss/base62"
	"io"
	"strings"
)

const infoPrefix = "node1"

type Info struct {
	Identity  id.Identity
	Alias     string
	Addresses []infra.Addr
}

func NewInfo(identity id.Identity) *Info {
	return &Info{
		Identity:  identity,
		Addresses: make([]infra.Addr, 0),
	}
}

func (info *Info) Add(addr infra.Addr) {
	for _, a := range info.Addresses {
		if infra.AddrEqual(a, addr) {
			return
		}
	}
	info.Addresses = append(info.Addresses, addr)
}

func (info Info) Pack() []byte {
	buf := &bytes.Buffer{}
	_ = writeInfo(buf, &info)
	return buf.Bytes()
}

func (info Info) String() string {
	return infoPrefix + base62.EncodeToString(info.Pack())
}

func Parse(s string) (*Info, error) {
	str := strings.TrimPrefix(s, infoPrefix)

	data, err := base62.DecodeString(str)
	if err != nil {
		return nil, err
	}

	return Unpack(data)
}

func Unpack(data []byte) (*Info, error) {
	return readInfo(bytes.NewReader(data))
}

func writeInfo(w io.Writer, info *Info) error {
	err := enc.WriteIdentity(w, info.Identity)
	if err != nil {
		return err
	}

	err = enc.WriteL8String(w, info.Alias)
	if err != nil {
		return err
	}

	addrs := info.Addresses[:]
	if len(addrs) > 255 {
		addrs = addrs[:255]
	}

	err = enc.Write(w, uint8(len(addrs)))
	if err != nil {
		return err
	}

	for _, addr := range addrs {
		if err := enc.WriteAddr(w, addr); err != nil {
			return nil
		}
	}

	return nil
}

func readInfo(r io.Reader) (*Info, error) {
	_id, err := enc.ReadIdentity(r)
	if err != nil {
		return nil, err
	}

	alias, err := enc.ReadL8String(r)
	if err != nil {
		return nil, err
	}

	count, err := enc.ReadUint8(r)
	if err != nil {
		return nil, err
	}

	addrs := make([]infra.Addr, 0, count)
	for i := 0; i < int(count); i++ {
		addr, err := enc.ReadAddr(r)
		if err != nil {
			return nil, err
		}
		addrs = append(addrs, addr)
	}

	return &Info{
		Identity:  _id,
		Alias:     alias,
		Addresses: addrs,
	}, nil
}
