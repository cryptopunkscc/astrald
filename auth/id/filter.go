package id

type Filter func(Identity) bool

var AllowEveryone = func(Identity) bool {
	return true
}

func AllowOnly(list ...Identity) Filter {
	var m map[string]struct{}
	for _, i := range list {
		m[i.PublicKeyHex()] = struct{}{}
	}
	return func(identity Identity) bool {
		_, ok := m[identity.PublicKeyHex()]
		return ok
	}
}

func AllowAny(filters ...Filter) Filter {
	return func(identity Identity) bool {
		for _, f := range filters {
			if f(identity) {
				return true
			}
		}
		return false
	}
}
