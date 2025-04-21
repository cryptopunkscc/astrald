package nodes

import (
	"errors"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/shell"
	"strings"
)

type opAddEndpointArgs struct {
	ID       *astral.Identity
	Endpoint string
}

func (mod *Module) OpAddEndpoint(_ *astral.Context, q shell.Query, args opAddEndpointArgs) (err error) {
	chunks := strings.SplitN(args.Endpoint, ":", 2)
	if len(chunks) != 2 {
		return errors.New("invalid endpoint")
	}
	parse, err := mod.Exonet.Parse(chunks[0], chunks[1])
	if err != nil {
		return
	}
	err = mod.AddEndpoint(args.ID, parse)
	if err != nil {
		return
	}
	_ = q.Accept().Close()
	return
}
