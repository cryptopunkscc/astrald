package astral

type Scope struct {
	QueryFilter IdentityFilter
}

func DefaultScope() *Scope {
	return &Scope{
		QueryFilter: nil,
	}
}
