package astral

import (
	"fmt"
	"strings"
)

// ErrRouteNotFound - failed to route the query to the destination
type ErrRouteNotFound struct {
	Router Router
	Fails  []error
}

var _ error = &ErrRouteNotFound{}

func NewErrRouteNotFound(router Router, errors ...error) *ErrRouteNotFound {
	return &ErrRouteNotFound{Router: router, Fails: errors}
}

func (e *ErrRouteNotFound) Error() string {
	errs := e.SubErrs()
	if len(errs) == 0 {
		return "route not found"
	}
	var s []string
	for _, err := range errs {
		s = append(s, err.Error())
	}

	return fmt.Sprintf("route not found: %v", strings.Join(s, ", "))
}

func (e *ErrRouteNotFound) SubErrs() (errs []error) {
	e.Walk(func(_ Router, err error) error {
		errs = append(errs, err)
		return nil
	})
	return
}

func (e *ErrRouteNotFound) Walk(fn func(Router, error) error) (err error) {
	for _, sub := range e.Fails {
		if rnf, ok := sub.(*ErrRouteNotFound); ok {
			err = rnf.Walk(fn)
			if err != nil {
				return
			}
			continue
		}
		err = fn(e.Router, sub)
		if err != nil {
			return
		}
	}

	return nil
}

func (e *ErrRouteNotFound) Is(other error) bool {
	_, ok := other.(*ErrRouteNotFound)
	return ok
}
