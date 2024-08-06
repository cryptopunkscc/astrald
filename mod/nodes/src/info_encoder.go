package nodes

import (
	"bytes"
	"errors"
	"github.com/cryptopunkscc/astrald/cslq"
)

type InfoEncoder struct {
	*Module
}

func (enc *InfoEncoder) Pack(info *NodeInfo) ([]byte, error) {
	var buf = &bytes.Buffer{}

	err := cslq.Encode(buf, "[c]c v c", info.Alias, info.Identity, len(info.Endpoints))
	if err != nil {
		return nil, err
	}

	for _, addr := range info.Endpoints {
		switch addr.Network() {
		case "tcp":
			if err := cslq.Encode(buf, "c", 0); err != nil {
				return nil, err
			}
		case "tor":
			if err := cslq.Encode(buf, "c", 1); err != nil {
				return nil, err
			}
		case "gw":
			if err := cslq.Encode(buf, "c", 2); err != nil {
				return nil, err
			}
		default:
			err := cslq.Encode(buf, "c[c]c", 255, addr.Network())
			if err != nil {
				return nil, err
			}
		}
		if err := cslq.Encode(buf, "[c]c", addr.Pack()); err != nil {
			return nil, err
		}
	}

	return buf.Bytes(), nil
}

func (enc *InfoEncoder) Unpack(data []byte) (*NodeInfo, error) {
	var count int
	var info NodeInfo
	var r = bytes.NewReader(data)

	if err := cslq.Decode(r, "[c]c v c", &info.Alias, info.Identity, &count); err != nil {
		return nil, err
	}

	for i := 0; i < count; i++ {
		var netCode int
		var netName string

		if err := cslq.Decode(r, "c", &netCode); err != nil {
			return nil, err
		}

		switch netCode {
		case 0:
			netName = "tcp"
		case 1:
			netName = "tor"
		case 2:
			netName = "gw"
		case 255:
			if err := cslq.Decode(r, "[c]c", &netName); err != nil {
				return nil, err
			}
		default:
			return nil, errors.New("unknown netcode")
		}

		var b []byte
		if err := cslq.Decode(r, "[c]c", &b); err != nil {
			return nil, err
		}

		e, err := enc.Exonet.Unpack(netName, b)
		if err != nil {
			return nil, err
		}

		info.Endpoints = append(info.Endpoints, e)
	}

	return &info, nil
}
