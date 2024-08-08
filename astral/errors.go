package astral

import (
	"errors"
	"fmt"
	"io"
	"strings"
)

// ErrRejected - the query was rejected by the target
//var ErrRejected = errors.New("query rejected")

type ErrRejected struct {
	Code uint8
}

func (e *ErrRejected) Error() string {
	return fmt.Sprintf("query rejected (%d)", e.Code)
}

func (e *ErrRejected) Is(other error) bool {
	_, ok := other.(*ErrRejected)
	return ok
}

// ErrAborted - query was aborted and routing did not finish
var ErrAborted = errors.New("query aborted")

// ErrTimeout - query timed out
var ErrTimeout = errors.New("query timeout")

// ErrZoneExcluded - operation requires zones excluded from the scope
var ErrZoneExcluded = errors.New("zone excluded")

// ErrRouteNotFound - failed to route the query to the destination
type ErrRouteNotFound struct {
	Router Router
	Fails  []error
}

func RouteNotFound(r Router, errors ...error) (io.WriteCloser, error) {
	return nil, &ErrRouteNotFound{
		Router: r,
		Fails:  errors,
	}
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
