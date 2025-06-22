package astral

// IdentityFilter returns false if the identity should be filtered out
type IdentityFilter func(*Identity) bool

// AllowEveryone always returns true
func AllowEveryone(*Identity) bool {
	return true
}

// AllowOnly returns an IdentityFilter that returns true only for identities on the list
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

// AllowAny returns an IdentityFilter that true if any of the provided filters returns true
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
