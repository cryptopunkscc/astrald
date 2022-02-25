package nodeinfo

import (
	"bytes"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/infra"
	"github.com/cryptopunkscc/astrald/infra/gw"
	"github.com/cryptopunkscc/astrald/infra/inet"
	"github.com/cryptopunkscc/astrald/infra/tor"
	"github.com/jxskiss/base62"
	"io"
	"strings"
)

const infoPrefix = "node1"

const netCodeInet = 0
const netCodeTor = 1
const netCodeGateway = 2

type Unpacker interface {
	Unpack(networkName string, addr []byte) (infra.Addr, error)
}

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

func Parse(s string, unpacker Unpacker) (*NodeInfo, error) {
	str := strings.TrimPrefix(s, infoPrefix)

	data, err := base62.DecodeString(str)
	if err != nil {
		return nil, err
	}

	return read(bytes.NewReader(data), unpacker)
}

func write(w io.Writer, c *NodeInfo) error {
	if err := cslq.WriteL8String(w, c.Alias); err != nil {
		return err
	}

	if err := cslq.WriteIdentity(w, c.Identity); err != nil {
		return err
	}

	addrs := c.Addresses[:]
	if len(addrs) > 255 {
		addrs = addrs[:255]
	}

	if err := cslq.Write(w, uint8(len(addrs))); err != nil {
		return err
	}

	for _, addr := range addrs {
		if err := writeAddr(w, addr); err != nil {
			return nil
		}
	}

	return nil
}

func read(r io.Reader, unpacker Unpacker) (*NodeInfo, error) {
	alias, err := cslq.ReadL8String(r)
	if err != nil {
		return nil, err
	}

	identity, err := cslq.ReadIdentity(r)
	if err != nil {
		return nil, err
	}

	count, err := cslq.ReadUint8(r)
	if err != nil {
		return nil, err
	}

	addrs := make([]infra.Addr, 0, count)
	for i := 0; i < int(count); i++ {
		addr, err := readAddr(r, unpacker)
		if err != nil {
			return nil, err
		}

		if gwAddr, ok := addr.(gw.Addr); ok {
			addr = gw.NewAddr(gwAddr.Gate(), identity.PublicKeyHex())
		}

		addrs = append(addrs, addr)
	}

	return &NodeInfo{
		Alias:     alias,
		Identity:  identity,
		Addresses: addrs,
	}, nil
}

func writeAddr(w io.Writer, addr infra.Addr) error {
	switch addr.Network() {
	case inet.NetworkName:
		if err := cslq.Write(w, uint8(netCodeInet)); err != nil {
			return err
		}
	case tor.NetworkName:
		if err := cslq.Write(w, uint8(netCodeTor)); err != nil {
			return err
		}
	case gw.NetworkName:
		if err := cslq.Write(w, uint8(netCodeGateway)); err != nil {
			return err
		}
		gwAddr, _ := addr.(gw.Addr)
		addr = gw.NewAddr(gwAddr.Gate(), "")
	default:
		if err := cslq.Write(w, uint8(255)); err != nil {
			return err
		}
		if err := cslq.WriteL8String(w, addr.Network()); err != nil {
			return err
		}
	}
	if err := cslq.WriteL8Bytes(w, addr.Pack()); err != nil {
		return err
	}
	return nil
}

func readAddr(r io.Reader, unpacker Unpacker) (infra.Addr, error) {
	net, err := cslq.ReadUint8(r)
	if err != nil {
		return nil, err
	}

	var netName string

	switch net {
	case netCodeInet:
		netName = inet.NetworkName
	case netCodeTor:
		netName = tor.NetworkName
	case netCodeGateway:
		netName = gw.NetworkName

	case 255:
		netName, err = cslq.ReadL8String(r)
		if err != nil {
			return nil, err
		}
	}

	data, err := cslq.ReadL8Bytes(r)
	if err != nil {
		return nil, err
	}

	addr, err := unpacker.Unpack(netName, data)

	return addr, err
}
