package bt

import (
	"github.com/cryptopunkscc/astrald/infra"
	"github.com/cryptopunkscc/astrald/infra/bt/bluez"
)

func (*Bluetooth) Addresses() []infra.AddrSpec {
	list := make([]infra.AddrSpec, 0)

	b, err := bluez.New()
	if err != nil {
		return list
	}

	adapters, err := b.Adapters()
	if err != nil {
		return list
	}

	for _, adapter := range adapters {
		if a, err := adapter.Address(); err == nil {
			if parsed, err := Parse(a); err == nil {
				list = append(list, infra.AddrSpec{
					Addr:   parsed,
					Global: false,
				})
			}
		}
	}

	return list
}
