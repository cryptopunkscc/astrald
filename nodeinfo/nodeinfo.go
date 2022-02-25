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

type Addr struct {
	Network string
	Data    []byte
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
	addrs := c.Addresses[:]
	if len(addrs) > 255 {
		addrs = addrs[:255]
	}

	err := cslq.Encode(w, "[c]c v c", c.Alias, c.Identity, len(addrs))
	if err != nil {
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
	var (
		alias    string
		identity id.Identity
		count    int
	)

	err := cslq.Decode(r, "[c]c v c", &alias, &identity, &count)
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
		if err := cslq.Encode(w, "c", netCodeInet); err != nil {
			return err
		}
	case tor.NetworkName:
		if err := cslq.Encode(w, "c", netCodeTor); err != nil {
			return err
		}
	case gw.NetworkName:
		if err := cslq.Encode(w, "c", netCodeGateway); err != nil {
			return err
		}
		gwAddr, _ := addr.(gw.Addr)
		addr = gw.NewAddr(gwAddr.Gate(), "")
	default:
		err := cslq.Encode(w, "c[c]c", 255, addr.Network())
		if err != nil {
			return err
		}
	}
	if err := cslq.Encode(w, "[c]c", addr.Pack()); err != nil {
		return err
	}
	return nil
}

func readAddr(r io.Reader, unpacker Unpacker) (infra.Addr, error) {
	var net int

	err := cslq.Decode(r, "c", &net)
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
		err = cslq.Decode(r, "[c]c", &netName)
		if err != nil {
			return nil, err
		}
	}

	var data []byte

	err = cslq.Decode(r, "[c]c", &data)
	if err != nil {
		return nil, err
	}

	addr, err := unpacker.Unpack(netName, data)

	return addr, err
}
