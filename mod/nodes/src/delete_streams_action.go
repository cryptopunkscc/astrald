package nodes

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/nodes"
)

type DeleteStreamsAction struct {
	mod *Module
}

func (mod *Module) NewDeleteStreamsAction() nodes.DeleteStreamAction {
	return &DeleteStreamsAction{mod: mod}
}

func (d DeleteStreamsAction) String() string {
	//TODO implement me
	panic("implement me")
}

func (d DeleteStreamsAction) Run(context *astral.Context) error {
	//TODO implement me
	panic("implement me")
}
