package astral

type Scope struct {
	Zone
	QueryFilter IdentityFilter
}

func DefaultScope() *Scope {
	return &Scope{
		Zone:        DefaultZones,
		QueryFilter: nil,
	}
}
