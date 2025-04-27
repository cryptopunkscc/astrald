package astral

import "io"

const (
	ZoneDevice = Zone(1 << iota)
	ZoneVirtual
	ZoneNetwork
)

const AllZones = ZoneDevice | ZoneVirtual | ZoneNetwork
const DefaultZones = ZoneDevice | ZoneVirtual

type Zone uint8

func (zone Zone) ObjectType() string {
	return "zone"
}

func (zone Zone) WriteTo(w io.Writer) (n int64, err error) {
	return Uint8(zone).WriteTo(w)
}

func (zone *Zone) ReadFrom(r io.Reader) (n int64, err error) {
	var u Uint8
	n, err = u.ReadFrom(r)
	*zone = Zone(u)
	return
}

// text support

func (zone Zone) MarshalText() (text []byte, err error) {
	return []byte(zone.String()), nil
}

func (zone *Zone) UnmarshalText(text []byte) error {
	*zone = Zones(string(text))
	return nil
}

func Zones(s string) (zone Zone) {
	for _, c := range s {
		switch c {
		case 'd':
			zone |= ZoneDevice
		case 'v':
			zone |= ZoneVirtual
		case 'n':
			zone |= ZoneNetwork
		}
	}
	return
}

func (zone Zone) Is(check Zone) bool {
	return zone&check == check
}

func (zone Zone) String() (s string) {
	if zone&ZoneDevice != 0 {
		s += "d"
	}
	if zone&ZoneVirtual != 0 {
		s += "v"
	}
	if zone&ZoneNetwork != 0 {
		s += "n"
	}
	return
}

func init() {
	var zone Zone
	DefaultBlueprints.Add(&zone)
}
