package astral

// IdentityFilter returns false if the identity should be filtered out
type IdentityFilter func(*Identity) bool

// AllowAll always returns true
func AllowAll(*Identity) bool {
	return true
}
