package user

import (
	"io"

	"github.com/cryptopunkscc/astrald/astral"
)

var _ astral.Object = &SwarmMember{}

type SwarmMember struct {
	Alias    astral.String
	Linked   astral.Bool
	Contract *NodeContract
}

func (s SwarmMember) ObjectType() string {
	return "mod.users.swarm_member"
}

func (s SwarmMember) WriteTo(w io.Writer) (n int64, err error) {
	o, err := astral.Objectify(s)
	if err != nil {
		return 0, err
	}

	return o.WriteTo(w)
}

func (s SwarmMember) ReadFrom(r io.Reader) (n int64, err error) {
	o, err := astral.Objectify(s)
	if err != nil {
		return 0, err
	}

	return o.ReadFrom(r)
}
