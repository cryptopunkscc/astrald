package auth

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/crypto"
)

const (
	ModuleName     = "auth"
	DBPrefix       = "auth__"
	OpSignContract = "auth.sign_contract"
	OpIndex        = "auth.index"
)

type ContractQueryBuilder interface {
	WithIssuer(*astral.Identity) ContractQueryBuilder
	WithSubject(*astral.Identity) ContractQueryBuilder
	WithAction(...astral.Object) ContractQueryBuilder
	Find(*astral.Context) ([]*SignedContract, error)
}

type Module interface {
	// Authorize checks whether the action is permitted.
	Authorize(ctx *astral.Context, action ActionObject) bool

	// Add registers typed handlers; the action type is inferred from each handler.
	Add(handlers ...TypedHandler)

	// VerifyContract verifies both signatures on a signed contract.
	VerifyContract(sc *SignedContract) error

	// SignIssuer signs the contract with the issuer's private key if needed.
	SignIssuer(ctx *astral.Context, contract *SignedContract) (*crypto.Signature, error)

	// SignSubject signs the contract with the subject's private key if needed.
	SignSubject(ctx *astral.Context, contract *SignedContract) (*crypto.Signature, error)

	// SignContract signs the contract with both issuer and subject private keys if needed.
	SignContract(ctx *astral.Context, contract *SignedContract) (*SignedContract, error)

	// IndexContract verifies and adds a signed contract to the auth index.
	IndexContract(ctx *astral.Context, contract *SignedContract) error

	// SignedContracts returns a query builder for finding active signed contracts.
	SignedContracts() ContractQueryBuilder
}
