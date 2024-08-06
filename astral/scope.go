package astral

type Zone int

type Scope struct {
	Zone
	QueryFilter IdentityFilter
}

const (
	ZoneDevice = Zone(1 << iota)
	ZoneVirtual
	ZoneNetwork
)

const AllZones = ZoneDevice | ZoneVirtual | ZoneNetwork
const DefaultZones = ZoneDevice | ZoneVirtual

func DefaultScope() *Scope {
	return &Scope{
		Zone:        DefaultZones,
		QueryFilter: nil,
	}
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
