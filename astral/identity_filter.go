package astral

type IdentityFilter func(*Identity) bool

var AllowEveryone = func(*Identity) bool {
	return true
}

func AllowOnly(list ...*Identity) IdentityFilter {
	var m map[string]struct{}
	for _, i := range list {
		m[i.String()] = struct{}{}
	}
	return func(identity *Identity) bool {
		_, ok := m[identity.String()]
		return ok
	}
}

func AllowAny(filters ...IdentityFilter) IdentityFilter {
	return func(identity *Identity) bool {
		for _, f := range filters {
			if f(identity) {
				return true
			}
		}
		return false
	}
}
