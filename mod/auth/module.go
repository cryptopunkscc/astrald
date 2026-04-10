package auth

import "github.com/cryptopunkscc/astrald/astral"

const (
	ModuleName = "auth"
	DBPrefix   = "auth__"
	ActionSudo = "mod.auth.sudo_action" // equals SudoAction{}.ObjectType()
)

type Module interface {
	// Authorize checks whether the action is permitted.
	// The action object carries the actor identity via Actor().
	// Returns true on first matching allow; false if no handler or contract allows.
	Authorize(ctx *astral.Context, action ActionObject) bool

	// Add registers one or more Handlers for a given action ObjectType string.
	// actionType must equal the ObjectType() of the action objects this handler
	// expects to receive.
	Add(actionType string, handlers ...Handler)
}
