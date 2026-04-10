package auth

import "github.com/cryptopunkscc/astrald/astral"

// Handler authorizes a typed action object.
// The action object carries the actor identity — no separate identity argument.
type Handler interface {
	Authorize(*astral.Context, ActionObject) bool
}

// Func is a generic adapter implementing Handler for a specific action type T.
// It type-asserts the incoming action object to T before dispatching.
// T must be an ActionObject (i.e. a concrete action type embedding auth.Action).
type Func[T ActionObject] func(*astral.Context, T) bool

func (f Func[T]) Authorize(ctx *astral.Context, action ActionObject) bool {
	if t, ok := action.(T); ok {
		return f(ctx, t)
	}
	return false
}
