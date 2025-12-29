package user

import (
	"io"

	"github.com/cryptopunkscc/astrald/astral"
)

var _ astral.Object = &SwarmMember{}

type SwarmMember struct {
	SignedContractID *astral.ObjectID
	Identity         *astral.Identity
	Alias            astral.String8
	Linked           astral.Bool
	Contract         *NodeContract
}

func (s SwarmMember) ObjectType() string {
	return "mod.users.swarm_member"
}

func (s SwarmMember) WriteTo(w io.Writer) (n int64, err error) {
	return astral.Objectify(&s).WriteTo(w)
}

func (s *SwarmMember) ReadFrom(r io.Reader) (n int64, err error) {
	return astral.Objectify(s).ReadFrom(r)
}

func init() {
	astral.DefaultBlueprints.Add(&SwarmMember{})
}
