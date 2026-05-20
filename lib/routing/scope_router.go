package routing

import (
	"errors"
	"io"
	"strings"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/lib/query"
	"github.com/cryptopunkscc/astrald/sig"
)

type ScopeRouter struct {
	root   astral.Router
	scopes sig.Map[string, astral.Router]
}

type HasSpec interface {
	Spec() (list []OpSpec)
}

type ScopedOpRouter interface {
	AddScopedOp(scope string, name string, op *Op) error
}

type RouteChecker interface {
	HasRoute(name string) bool
}

func NewScopeRouter(root astral.Router) *ScopeRouter {
	if root == nil {
		root = &NilRouter{}
	}
	return &ScopeRouter{
		root: root,
	}
}

func (r *ScopeRouter) RouteQuery(ctx *astral.Context, q *astral.InFlightQuery, w io.WriteCloser) (io.WriteCloser, error) {
	opName, _ := query.Parse(q.QueryString)
	idx := strings.IndexByte(opName, '.')
	if idx == -1 {
		return r.root.RouteQuery(ctx, q, w)
	}
	scopeName, opName := opName[:idx], opName[idx+1:]

	scope, ok := r.scopes.Get(scopeName)
	if !ok {
		return r.root.RouteQuery(ctx, q, w)
	}

	// remove the prefix from the query
	rq := astral.Launch(&astral.Query{
		Nonce:       q.Query.Nonce,
		Caller:      q.Query.Caller,
		Target:      q.Query.Target,
		QueryString: q.QueryString[idx+1:],
	})

	rq.Extra = q.Extra

	return scope.RouteQuery(ctx, rq, w)
}

func (r *ScopeRouter) Add(scope string, router astral.Router) {
	r.scopes.Set(scope, router)
}

func (r *ScopeRouter) AddScopedOp(scope string, name string, op *Op) error {
	if scope == "" {
		root, ok := r.root.(*OpRouter)
		if !ok {
			return errors.New("root router is not an op router")
		}
		return root.AddOp(name, op)
	}

	router, ok := r.scopes.Get(scope)
	if !ok {
		router = NewOpRouter()
		r.scopes.Set(scope, router)
	}

	ops, ok := router.(*OpRouter)
	if !ok {
		return errors.New("scope router is not an op router")
	}
	return ops.AddOp(name, op)
}

func (r *ScopeRouter) HasRoute(name string) bool {
	idx := strings.IndexByte(name, '.')
	if idx == -1 {
		root, ok := r.root.(RouteChecker)
		return ok && root.HasRoute(name)
	}

	scopeName, opName := name[:idx], name[idx+1:]
	scope, ok := r.scopes.Get(scopeName)
	if ok {
		checker, ok := scope.(RouteChecker)
		return ok && checker.HasRoute(opName)
	}

	root, ok := r.root.(RouteChecker)
	return ok && root.HasRoute(name)
}

func (r *ScopeRouter) Remove(scope string) {
	r.scopes.Delete(scope)
}

func (r *ScopeRouter) Spec() (list []OpSpec) {
	for name, scope := range r.scopes.Clone() {
		r, ok := scope.(HasSpec)
		if !ok {
			continue
		}

		subList := r.Spec()
		for _, opSpec := range subList {
			opSpec.Name = name + "." + opSpec.Name
			list = append(list, opSpec)
		}
	}

	if r, ok := r.root.(HasSpec); ok {
		for _, opSpec := range r.Spec() {
			if strings.HasPrefix(opSpec.Name, ".") {
				continue
			}
			list = append(list, opSpec)
		}
	}
	return
}
