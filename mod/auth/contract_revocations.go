package auth

import (
	"io"

	"github.com/cryptopunkscc/astrald/astral"
)

// ContractRevocations is a list of contracts revoked by a single issuer.
type ContractRevocations struct {
	ContractID *astral.ObjectID
	ExpiresAt  astral.Time // Required to have as we could not posses contract anymore
	CreatedAt  astral.Time // purely informational
}

var _ astral.Object = &ContractRevocations{}

func (ContractRevocations) ObjectType() string { return "mod.auth.contract_revocations" }

func (r ContractRevocations) WriteTo(w io.Writer) (int64, error) {
	return astral.Objectify(&r).WriteTo(w)
}

func (r *ContractRevocations) ReadFrom(rd io.Reader) (int64, error) {
	return astral.Objectify(r).ReadFrom(rd)
}

func (r ContractRevocations) MarshalJSON() ([]byte, error) {
	return astral.Objectify(&r).MarshalJSON()
}

func (r *ContractRevocations) UnmarshalJSON(b []byte) error {
	return astral.Objectify(r).UnmarshalJSON(b)
}

func init() { _ = astral.Add(&ContractRevocations{}) }
