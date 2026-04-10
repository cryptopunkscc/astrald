package auth

import "github.com/cryptopunkscc/astrald/astral"

const (
	ModuleName = "auth"
	DBPrefix   = "auth__"
	ActionSudo = "mod.auth.sudo_action" // equals SudoAction{}.ObjectType()
)

type Module interface {
	// Authorize checks whether the action is permitted.
	Authorize(ctx *astral.Context, action ActionObject) bool
	Add(actionType string, handlers ...Handler)

	// VerifyContract verifies both signatures on a signed contract.
	VerifyContract(sc *SignedContract) error

	SignContract(ctx *astral.Context, contract *Contract) (*SignedContract, error)
	// IndexContract loads a signed contract from the object store and adds it to the auth index.
	IndexContract(ctx *astral.Context, objectID *astral.ObjectID) error

	// StoreContract saves a signed contract to the object store and indexes it.
	StoreContract(ctx *astral.Context, sc *SignedContract) error

	// FindContractsWithActor returns active contracts where identity is the Subject.
	FindContractsWithActor(ctx *astral.Context, actor *astral.Identity) ([]*SignedContract, error)

	// FindContractsWithIssuer returns active contracts where identity is the Issuer.
	FindContractsWithIssuer(ctx *astral.Context, issuer *astral.Identity) ([]*SignedContract, error)

	Ban(ctx *astral.Context, identity *astral.Identity) error
	// IsBanned reports whether an identity is currently banned.
	IsBanned(identity *astral.Identity) bool
}
