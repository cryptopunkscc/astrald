package infra

import (
	"errors"
	"github.com/cryptopunkscc/astrald/infra"
	"github.com/cryptopunkscc/astrald/infra/bt"
	"github.com/cryptopunkscc/astrald/infra/gw"
	"github.com/cryptopunkscc/astrald/infra/inet"
	"github.com/cryptopunkscc/astrald/infra/tor"
)

func (i *Infra) Unpack(network string, data []byte) (infra.Addr, error) {
	switch network {
	case inet.NetworkName:
		return inet.Unpack(data)
	case tor.NetworkName:
		return tor.Unpack(data)
	case gw.NetworkName:
		return gw.Unpack(data)
	case bt.NetworkName:
		return bt.Unpack(data)
	}

	return infra.NewGenericAddr(network, data), nil
}

func (i *Infra) Parse(network string, hexBytes string) (infra.Addr, error) {
	switch network {
	case inet.NetworkName:
		return inet.Parse(hexBytes)
	case tor.NetworkName:
		return tor.Parse(hexBytes)
	case gw.NetworkName:
		return gw.Parse(hexBytes)
	case bt.NetworkName:
		return bt.Parse(hexBytes)
	}

	return nil, errors.New("unsupported network")
}
