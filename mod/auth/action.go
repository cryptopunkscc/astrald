package auth

import (
	"github.com/cryptopunkscc/astrald/astral"
	"io"
)

type Action string

func (a Action) ObjectType() string {
	return "astrald.mod.auth.action"
}

func (a Action) WriteTo(w io.Writer) (n int64, err error) {
	return astral.String8(a).WriteTo(w)
}

func (a *Action) ReadFrom(r io.Reader) (n int64, err error) {
	return (*astral.String8)(a).ReadFrom(r)
}

func (a Action) String() string {
	return string(a)
}
