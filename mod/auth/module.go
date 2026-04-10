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

	// VerifyContract verifies both signatures on a signed contract.
	VerifyContract(sc *SignedContract) error

	// SignContract signs both the Issuer and Subject sides of a contract using
	// locally available keys. Tries ASN1 first, falls back to BIP137.
	SignContract(ctx *astral.Context, contract *Contract) (*SignedContract, error)
	// StoreContract saves a signed contract to the object store and indexes it.
	StoreContract(ctx *astral.Context, sc *SignedContract) error
	// FindContractsWithActor returns active contracts where identity is the Subject.
	FindContractsWithActor(ctx *astral.Context, actor *astral.Identity) ([]*SignedContract, error)

	// FindContractsWithIssuer returns active contracts where identity is the Issuer.
	FindContractsWithIssuer(ctx *astral.Context, issuer *astral.Identity) ([]*SignedContract, error)

	// Ban marks an identity as banned. Banned identities are denied all actions.
	Ban(ctx *astral.Context, identity *astral.Identity) error
	// IsBanned reports whether an identity is currently banned.
	IsBanned(identity *astral.Identity) bool
}
