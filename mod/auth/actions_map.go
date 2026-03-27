package auth

import (
	"github.com/cryptopunkscc/astrald/astral"
)

type Handler func(*astral.Context, *astral.Identity, astral.Object) bool

// NewHandler returns a Handler that calls fn with the context, identity and type-checked target.
func NewHandler[T astral.Object](fn func(*astral.Context, *astral.Identity, T) bool) Handler {
	return func(ctx *astral.Context, identity *astral.Identity, target astral.Object) bool {
		t, ok := target.(T)
		if !ok {
			return false
		}
		return fn(ctx, identity, t)
	}
}

// AddAuthorizer wraps each fn as a type-checked Handler and registers them for action on m.
func AddAuthorizer[T astral.Object](m Module, action Action, fns ...func(*astral.Context, *astral.Identity, T) bool) {
	handlers := make([]Handler, len(fns))
	for i, fn := range fns {
		handlers[i] = NewHandler(fn)
	}
	m.AddAuthorizer(action, handlers...)
}
