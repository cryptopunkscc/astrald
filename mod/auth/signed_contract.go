package auth

import (
	"io"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/crypto"
)

type SignedContract struct {
	*Contract
	IssuerSig *crypto.Signature
	SubjecSig *crypto.Signature
}

var _ astral.Object = &SignedContract{}

func (SignedContract) ObjectType() string { return "mod.auth.signed_contract" }

func (c SignedContract) WriteTo(w io.Writer) (int64, error)   { return astral.Objectify(&c).WriteTo(w) }
func (c *SignedContract) ReadFrom(r io.Reader) (int64, error) { return astral.Objectify(c).ReadFrom(r) }

func (c SignedContract) MarshalJSON() ([]byte, error)  { return astral.Objectify(&c).MarshalJSON() }
func (c *SignedContract) UnmarshalJSON(b []byte) error { return astral.Objectify(c).UnmarshalJSON(b) }

func (c *SignedContract) IsNil() bool { return c == nil || c.Contract == nil }

func init() { _ = astral.Add(&SignedContract{}) }
