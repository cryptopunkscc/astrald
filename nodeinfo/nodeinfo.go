package nodeinfo

import (
	"bytes"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/enc"
	"github.com/cryptopunkscc/astrald/infra"
	"github.com/jxskiss/base62"
	"io"
	"strings"
)

const infoPrefix = "node1"

type NodeInfo struct {
	Alias     string
	Identity  id.Identity
	Addresses []infra.Addr
}

func New(identity id.Identity) *NodeInfo {
	return &NodeInfo{
		Identity:  identity,
		Addresses: make([]infra.Addr, 0),
	}
}

func (info *NodeInfo) String() string {
	buf := &bytes.Buffer{}
	write(buf, info)
	return infoPrefix + base62.EncodeToString(buf.Bytes())
}

func Parse(s string) (*NodeInfo, error) {
	str := strings.TrimPrefix(s, infoPrefix)

	data, err := base62.DecodeString(str)
	if err != nil {
		return nil, err
	}

	return read(bytes.NewReader(data))
}

func write(w io.Writer, c *NodeInfo) error {
	if err := enc.WriteL8String(w, c.Alias); err != nil {
		return err
	}

	if err := enc.WriteIdentity(w, c.Identity); err != nil {
		return err
	}

	addrs := c.Addresses[:]
	if len(addrs) > 255 {
		addrs = addrs[:255]
	}

	if err := enc.Write(w, uint8(len(addrs))); err != nil {
		return err
	}

	for _, addr := range addrs {
		if err := WriteAddr(w, addr); err != nil {
			return nil
		}
	}

	return nil
}

func read(r io.Reader) (*NodeInfo, error) {
	alias, err := enc.ReadL8String(r)
	if err != nil {
		return nil, err
	}

	identity, err := enc.ReadIdentity(r)
	if err != nil {
		return nil, err
	}

	count, err := enc.ReadUint8(r)
	if err != nil {
		return nil, err
	}

	addrs := make([]infra.Addr, 0, count)
	for i := 0; i < int(count); i++ {
		addr, err := ReadAddr(r)
		if err != nil {
			return nil, err
		}
		addrs = append(addrs, addr)
	}

	return &NodeInfo{
		Alias:     alias,
		Identity:  identity,
		Addresses: addrs,
	}, nil
}

func WriteAddr(w io.Writer, addr infra.Addr) error {
	switch addr.Network() {
	case "inet":
		if err := enc.Write(w, uint8(0)); err != nil {
			return err
		}
	case "tor":
		if err := enc.Write(w, uint8(1)); err != nil {
			return err
		}
	default:
		if err := enc.Write(w, uint8(255)); err != nil {
			return err
		}
		if err := enc.WriteL8String(w, addr.Network()); err != nil {
			return err
		}
	}
	if err := enc.WriteL8Bytes(w, addr.Pack()); err != nil {
		return err
	}
	return nil
}

func ReadAddr(r io.Reader) (infra.Addr, error) {
	net, err := enc.ReadUint8(r)
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
		netName, err = enc.ReadL8String(r)
		if err != nil {
			return nil, err
		}
	}

	data, err := enc.ReadL8Bytes(r)
	if err != nil {
		return nil, err
	}

	return astral.Unpack(netName, data)
}
