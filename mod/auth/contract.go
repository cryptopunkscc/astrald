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
	Permits   *astral.Slice[*Permit]
	ExpiresAt astral.Time
}

type Permit struct {
	Action      astral.String8 // object type of an action
	Constraints *astral.Bundle // list of constraints
}

var _ astral.Object = &Permit{}
var _ crypto.SignableTextObject = &Contract{}

func (Permit) ObjectType() string { return "mod.auth.permit" }

func (p Permit) WriteTo(w io.Writer) (int64, error)   { return astral.Objectify(&p).WriteTo(w) }
func (p *Permit) ReadFrom(r io.Reader) (int64, error) { return astral.Objectify(p).ReadFrom(r) }

func (p Permit) MarshalJSON() ([]byte, error)  { return astral.Objectify(&p).MarshalJSON() }
func (p *Permit) UnmarshalJSON(b []byte) error { return astral.Objectify(p).UnmarshalJSON(b) }

func (c *Contract) Allows(action ActionObject) bool {
	if c.Permits == nil {
		return false
	}
	for _, p := range *c.Permits.Elem {
		if ca, ok := action.(Constrainable); ok {
			if !ca.ApplyConstraints(p.Constraints) {
				continue
			}
		}
		return true
	}
	return false
}

// HasPermit returns all permits in this contract that match the given action type.
// Empty result means the contract grants no such permission.
func (c *Contract) HasPermit(action string) []*Permit {
	if c.Permits == nil {
		return nil
	}
	var result []*Permit
	for _, p := range *c.Permits.Elem {
		if string(p.Action) == action {
			result = append(result, p)
		}
	}
	return result
}

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

func init() {
	_ = astral.Add(&Contract{})
	_ = astral.Add(&Permit{})
}
