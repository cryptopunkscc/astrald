package auth

import (
	"github.com/cryptopunkscc/astrald/astral"
)

// Action is the base struct embedded by all typed action objects.
type Action struct {
	Nonce   astral.Nonce
	ActorId *astral.Identity
}

// NewAction returns an Action with a fresh nonce and the given actor.
func NewAction(actor *astral.Identity) Action {
	return Action{Nonce: astral.NewNonce(), ActorId: actor}
}

func (a Action) Id() astral.Nonce              { return a.Nonce }
func (a Action) Actor() *astral.Identity       { return a.ActorId }
func (a *Action) SetActor(id *astral.Identity) { a.ActorId = id }

// ActionObject is the interface satisfied by all action types.
type ActionObject interface {
	astral.Object
	Id() astral.Nonce
	Actor() *astral.Identity
	SetActor(*astral.Identity)
}

// Constrainable is implemented by actions that know how to evaluate permit constraints.
// Actions that do NOT implement this interface are always permitted regardless of constraints.
type Constrainable interface {
	ApplyConstraints(*astral.Bundle) bool
}
