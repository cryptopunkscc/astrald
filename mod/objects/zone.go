package objects

type Zone int

const (
	ZoneLocal = Zone(1 << iota)
	ZoneVirtual
	ZoneNetwork
)

const DefaultZones = ZoneLocal | ZoneVirtual

func Zones(s string) (zone Zone) {
	for _, c := range s {
		switch c {
		case 'l':
			zone |= ZoneLocal
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
	if zone&ZoneLocal != 0 {
		s += "l"
	}
	if zone&ZoneVirtual != 0 {
		s += "v"
	}
	if zone&ZoneNetwork != 0 {
		s += "n"
	}
	return
}
