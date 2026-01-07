package user

import (
	"io"

	"github.com/cryptopunkscc/astrald/astral"
)

// Info contains information about user's node configuration
type Info struct {
	NodeAlias  astral.String8
	UserAlias  astral.String8
	ContractID *astral.ObjectID

	Contract *SignedNodeContract
}

func (i Info) ObjectType() string { return "mod.user.info" }

func (i Info) WriteTo(w io.Writer) (n int64, err error) {
	return astral.Struct(i).WriteTo(w)
}

func (i *Info) ReadFrom(r io.Reader) (n int64, err error) {
	return astral.Struct(i).ReadFrom(r)
}

func init() {
	astral.DefaultBlueprints.Add(&Info{})
}
