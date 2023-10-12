package nodeinfo

import (
	"errors"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/node/infra/drivers/inet"
	"github.com/cryptopunkscc/astrald/node/infra/drivers/tor"
)

const infoPrefix = "node1"

const netCodeInet = 0
const netCodeTor = 1
const netCodeGateway = 2
const netCodeOther = 255
const netNameGateway = "gw"

func (info *NodeInfo) UnmarshalCSLQ(dec *cslq.Decoder) error {
	var count int

	if err := dec.Decodef("[c]c v c", &info.Alias, &info.Identity, &count); err != nil {
		return err
	}

	addrs := make([]net.Endpoint, 0, count)
	for i := 0; i < count; i++ {
		addr, err := decodeAddr(dec)
		if err != nil {
			return err
		}
		addrs = append(addrs, addr)
	}
	info.Endpoints = addrs

	return nil
}

func (info *NodeInfo) MarshalCSLQ(enc *cslq.Encoder) error {
	endpoints := info.Endpoints
	if len(endpoints) > 255 {
		endpoints = endpoints[:255]
	}

	err := enc.Encodef("[c]c v c", info.Alias, info.Identity, len(endpoints))
	if err != nil {
		return err
	}

	for _, addr := range endpoints {
		if err := encodeAddr(enc, addr); err != nil {
			return nil
		}
	}

	return nil
}

func encodeAddr(enc *cslq.Encoder, addr net.Endpoint) error {
	switch addr.Network() {
	case inet.DriverName:
		if err := enc.Encodef("c", netCodeInet); err != nil {
			return err
		}
	case tor.DriverName:
		if err := enc.Encodef("c", netCodeTor); err != nil {
			return err
		}
	case netNameGateway:
		if err := enc.Encodef("c", netCodeGateway); err != nil {
			return err
		}
	default:
		err := enc.Encodef("c[c]c", 255, addr.Network())
		if err != nil {
			return err
		}
	}
	if err := enc.Encodef("[c]c", addr.Pack()); err != nil {
		return err
	}
	return nil
}

func decodeAddr(dec *cslq.Decoder) (*net.GenericEndpoint, error) {
	var netCode int
	var netName string

	if err := dec.Decodef("c", &netCode); err != nil {
		return nil, err
	}

	switch netCode {
	case netCodeInet:
		netName = inet.DriverName

	case netCodeTor:
		netName = tor.DriverName

	case netCodeGateway:
		netName = netNameGateway

	case netCodeOther:
		if err := dec.Decodef("[c]c", &netName); err != nil {
			return nil, err
		}
	default:
		return nil, errors.New("unknown netcode")
	}

	var data []byte
	if err := dec.Decodef("[c]c", &data); err != nil {
		return nil, err
	}

	return net.NewGenericEndpoint(netName, data), nil
}
