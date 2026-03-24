package auth

import (
	"reflect"

	"github.com/cryptopunkscc/astrald/astral"
)

// Handler is a pre-compiled authorization handler for a specific target type.
type Handler struct {
	argType reflect.Type
	fn      reflect.Value
}

// NewHandler wraps a typed handler function into an Handler.
// The handler receives a strongly-typed target; the type assertion is performed by Call.
func NewHandler[T astral.Object](fn func(*astral.Identity, T) bool) Handler {
	return Handler{
		argType: reflect.TypeOf((*T)(nil)).Elem(),
		fn:      reflect.ValueOf(fn),
	}
}

// Call invokes the handler if target is assignable to the handler's expected type.
// Returns false if target is nil or the type doesn't match.
func (h Handler) Call(identity *astral.Identity, target astral.Object) bool {
	if target == nil {
		return false
	}
	targetVal := reflect.ValueOf(target)
	if !targetVal.Type().AssignableTo(h.argType) {
		return false
	}
	result := h.fn.Call([]reflect.Value{reflect.ValueOf(identity), targetVal})
	return result[0].Bool()
}

// ActionsMap maps Actions to a list of AuthHandlers.
type ActionsMap map[Action][]Handler

// Auth dispatches an authorization request to the matching handlers in m.
// Returns true on the first handler that returns true.
func Auth(m ActionsMap, identity *astral.Identity, action Action, target astral.Object) bool {
	for _, h := range m[action] {
		if h.Call(identity, target) {
			return true
		}
	}
	return false
}
