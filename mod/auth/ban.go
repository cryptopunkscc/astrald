package auth

import (
	"io"

	"github.com/cryptopunkscc/astrald/astral"
)

type Ban struct {
	SubjectID *astral.Identity
	CreatedAt astral.Time
}

var _ astral.Object = &Ban{}

func (Ban) ObjectType() string { return "mod.auth.ban" }

func (b Ban) WriteTo(w io.Writer) (int64, error)   { return astral.Objectify(&b).WriteTo(w) }
func (b *Ban) ReadFrom(r io.Reader) (int64, error) { return astral.Objectify(b).ReadFrom(r) }

func init() { _ = astral.Add(&Ban{}) }
