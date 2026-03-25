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

type ActionsMap map[Action][]Handler

func Auth(m ActionsMap, ctx *astral.Context, identity *astral.Identity, action Action, target astral.Object) bool {
	for _, h := range m[action] {
		if h(ctx, identity, target) {
			return true
		}
	}
	return false
}
