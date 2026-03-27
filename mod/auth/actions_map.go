package auth

import (
	"github.com/cryptopunkscc/astrald/astral"
)

// Handler is implemented by any type that can authorize an action.
type Handler interface {
	Authorize(*astral.Context, *astral.Identity, astral.Object) bool
}

// Func is a generic named function type implementing Handler with automatic type checking.
type Func[T astral.Object] func(*astral.Context, *astral.Identity, T) bool

func (f Func[T]) Authorize(ctx *astral.Context, identity *astral.Identity, target astral.Object) bool {
	if t, ok := target.(T); ok {
		return f(ctx, identity, t)
	}
	return false
}
