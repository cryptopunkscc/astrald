package auth

import (
	"github.com/cryptopunkscc/astrald/astral"
)

type Handler func(*astral.Identity, astral.Object) bool

// NewHandler returns a Handler that calls fn with the identity and type-checked target.
func NewHandler[T astral.Object](fn func(*astral.Identity, T) bool) Handler {
	return func(identity *astral.Identity, target astral.Object) bool {
		t, ok := target.(T)
		if !ok {
			return false
		}
		return fn(identity, t)
	}
}

type ActionsMap map[Action][]Handler

func Auth(m ActionsMap, identity *astral.Identity, action Action, target astral.Object) bool {
	for _, h := range m[action] {
		if h(identity, target) {
			return true
		}
	}
	return false
}
