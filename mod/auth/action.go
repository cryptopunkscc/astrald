package auth

import "github.com/cryptopunkscc/astrald/astral"

// Action is the base struct embedded by all typed action objects.
type Action struct {
	Id      astral.Nonce
	ActorId *astral.Identity
}

// NewAction returns an Action with a fresh nonce and the given actor.
func NewAction(actor *astral.Identity) Action {
	return Action{Id: astral.NewNonce(), ActorId: actor}
}

// Actor returns the identity of the actor making the request.
// Promoted to all concrete action types that embed Action.
func (a Action) Actor() *astral.Identity {
	return a.ActorId
}

// ActorAction is implemented by any action that carries an actor identity.
// All concrete action types that embed Action satisfy this automatically
// via method promotion.
type ActorAction interface {
	Actor() *astral.Identity
}

// ActionObject is the interface satisfied by all action types.
// Embedding auth.Action satisfies Actor() via promotion; the concrete type
// must additionally implement astral.Object (ObjectType, WriteTo, ReadFrom).
type ActionObject interface {
	astral.Object
	Actor() *astral.Identity
}

// Constrainable is implemented by actions that know how to evaluate permit constraints.
// Actions that do NOT implement this interface are always permitted regardless of constraints.
type Constrainable interface {
	ApplyConstraints([]Constraint) bool
}
