package astral

// ErrRouteNotFound - the target was unreachable via the router
type ErrRouteNotFound struct{}

var _ error = &ErrRouteNotFound{}

func NewErrRouteNotFound() *ErrRouteNotFound {
	return &ErrRouteNotFound{}
}

func (e *ErrRouteNotFound) Error() string {
	return "route not found"
}

func (e *ErrRouteNotFound) Is(other error) bool {
	_, ok := other.(*ErrRouteNotFound)
	return ok
}
