package user

import (
	"github.com/cryptopunkscc/astrald/astral"
	"io"
)

// Info contains information about user's node configuration
type Info struct {
	NodeAlias astral.String8
	UserAlias astral.String8
}

func (i Info) ObjectType() string { return "mod.user.info" }

func (i Info) WriteTo(w io.Writer) (n int64, err error) {
	return astral.Struct(i).WriteTo(w)
}

func (i *Info) ReadFrom(r io.Reader) (n int64, err error) {
	return astral.Struct(i).ReadFrom(r)
}
