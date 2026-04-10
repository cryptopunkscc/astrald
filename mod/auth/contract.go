package auth

import (
	"fmt"
	"io"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/crypto"
)

type Contract struct {
	Issuer    *astral.Identity
	Subject   *astral.Identity
	Permits   []Permit
	ExpiresAt astral.Time
}

type Permit struct {
	Action      astral.String8  // object type of an action
	Constraints []astral.Object // list of constraints
}

var _ crypto.SignableTextObject = &Contract{}

func (Contract) ObjectType() string { return "mod.auth.contract" }

func (c Contract) WriteTo(w io.Writer) (int64, error)   { return astral.Objectify(&c).WriteTo(w) }
func (c *Contract) ReadFrom(r io.Reader) (int64, error) { return astral.Objectify(c).ReadFrom(r) }

func (c Contract) MarshalJSON() ([]byte, error)  { return astral.Objectify(&c).MarshalJSON() }
func (c *Contract) UnmarshalJSON(b []byte) error { return astral.Objectify(c).UnmarshalJSON(b) }

func (c *Contract) SignableHash() []byte {
	id, err := astral.ResolveObjectID(c)
	if err != nil {
		return nil
	}
	return id.Hash[:]
}

func (c *Contract) SignableText() string {
	return fmt.Sprintf(
		"%s grants %s permissions until %s",
		c.Issuer.String(),
		c.Subject.String(),
		c.ExpiresAt.Time().Format("2006-01-02 15:04:05"),
	)
}

func init() { _ = astral.Add(&Contract{}) }
