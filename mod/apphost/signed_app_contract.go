package apphost

import (
	"io"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/crypto"
)

type SignedAppContract struct {
	*AppContract
	AppSig  *crypto.Signature
	HostSig *crypto.Signature
}

var _ astral.Object = &SignedAppContract{}

func (SignedAppContract) ObjectType() string { return "mod.apphost.signed_app_contract" }

func (c SignedAppContract) WriteTo(w io.Writer) (n int64, err error) {
	return astral.Objectify(&c).WriteTo(w)
}

func (c *SignedAppContract) ReadFrom(r io.Reader) (n int64, err error) {
	return astral.Objectify(c).ReadFrom(r)
}

func (c SignedAppContract) MarshalJSON() ([]byte, error) {
	return astral.Objectify(&c).MarshalJSON()
}

func (c *SignedAppContract) UnmarshalJSON(bytes []byte) error {
	return astral.Objectify(c).UnmarshalJSON(bytes)
}

func (c *SignedAppContract) IsNil() bool { return c == nil || c.AppContract == nil }

func init() {
	astral.Add(&SignedAppContract{})
}
