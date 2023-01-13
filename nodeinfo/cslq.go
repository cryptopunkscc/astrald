package nodeinfo

import (
	"errors"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/infra"
	"github.com/cryptopunkscc/astrald/infra/gw"
	"github.com/cryptopunkscc/astrald/infra/inet"
	"github.com/cryptopunkscc/astrald/infra/tor"
)

const infoPrefix = "node1"

const netCodeInet = 0
const netCodeTor = 1
const netCodeGateway = 2
const netCodeOther = 255

func (info *NodeInfo) UnmarshalCSLQ(dec *cslq.Decoder) error {
	var count int

	if err := dec.Decode("[c]c v c", &info.Alias, &info.Identity, &count); err != nil {
		return err
	}

	addrs := make([]infra.Addr, 0, count)
	for i := 0; i < count; i++ {
		addr, err := decodeAddr(dec)
		if err != nil {
			return err
		}
		addrs = append(addrs, addr)
	}
	info.Addresses = addrs

	return nil
}

func (info *NodeInfo) MarshalCSLQ(enc *cslq.Encoder) error {
	addrs := info.Addresses[:]
	if len(addrs) > 255 {
		addrs = addrs[:255]
	}

	err := enc.Encode("[c]c v c", info.Alias, info.Identity, len(addrs))
	if err != nil {
		return err
	}

	for _, addr := range addrs {
		if err := encodeAddr(enc, addr); err != nil {
			return nil
		}
	}

	return nil
}

func encodeAddr(enc *cslq.Encoder, addr infra.Addr) error {
	switch addr.Network() {
	case inet.NetworkName:
		if err := enc.Encode("c", netCodeInet); err != nil {
			return err
		}
	case tor.NetworkName:
		if err := enc.Encode("c", netCodeTor); err != nil {
			return err
		}
	case gw.NetworkName:
		if err := enc.Encode("c", netCodeGateway); err != nil {
			return err
		}
		gwAddr, _ := addr.(gw.Addr)
		addr = gw.NewAddr(gwAddr.Gate(), gwAddr.Target())
	default:
		err := enc.Encode("c[c]c", 255, addr.Network())
		if err != nil {
			return err
		}
	}
	if err := enc.Encode("[c]c", addr.Pack()); err != nil {
		return err
	}
	return nil
}

func decodeAddr(dec *cslq.Decoder) (*infra.GenericAddr, error) {
	var netCode int
	var netName string

	if err := dec.Decode("c", &netCode); err != nil {
		return nil, err
	}

	switch netCode {
	case netCodeInet:
		netName = inet.NetworkName

	case netCodeTor:
		netName = tor.NetworkName

	case netCodeGateway:
		netName = gw.NetworkName

	case netCodeOther:
		if err := dec.Decode("[c]c", &netName); err != nil {
			return nil, err
		}
	default:
		return nil, errors.New("unknown netcode")
	}

	var data []byte
	if err := dec.Decode("[c]c", &data); err != nil {
		return nil, err
	}

	return infra.NewGenericAddr(netName, data), nil
}
