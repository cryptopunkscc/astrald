package apphost

import (
	"github.com/cryptopunkscc/astrald/astral"
	"io"
)

const ModuleName = "apphost"
const DBPrefix = "apphost__"

type Module interface {
}

type AccessToken struct {
	Identity  *astral.Identity
	Token     astral.String8
	ExpiresAt astral.Time
}

var _ astral.Object = &AccessToken{}

func (at AccessToken) ObjectType() string { return "astrald.mod.apphost.access_token" }

func (at AccessToken) WriteTo(w io.Writer) (n int64, err error) {
	return astral.Struct(at).WriteTo(w)
}

func (at *AccessToken) ReadFrom(r io.Reader) (n int64, err error) {
	return astral.Struct(at).ReadFrom(r)
}
