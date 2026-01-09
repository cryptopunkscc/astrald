package user

import (
	"github.com/cryptopunkscc/astrald/astral"
	"io"
)

type Notification struct {
	Event astral.String8
}

var _ astral.Object = &Notification{}

func (n Notification) ObjectType() string {
	return "mod.user.notification"
}

func (n Notification) WriteTo(w io.Writer) (int64, error) {
	return astral.Struct(n).WriteTo(w)
}

func (n *Notification) ReadFrom(r io.Reader) (int64, error) {
	return astral.Struct(n).ReadFrom(r)
}

func init() {
	astral.Add(&Notification{})
}
