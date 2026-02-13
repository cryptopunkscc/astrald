package nodes

import (
	"errors"
	"strings"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/ops"
)

type opAddEndpointArgs struct {
	ID       *astral.Identity
	Endpoint string
	In       string `query:"optional"`
	Out      string `query:"optional"`
}

func (mod *Module) OpAddEndpoint(_ *astral.Context, q *ops.Query, args opAddEndpointArgs) (err error) {
	chunks := strings.SplitN(args.Endpoint, ":", 2)
	if len(chunks) != 2 {
		return errors.New("invalid endpoint")
	}

	ch := channel.New(q.Accept(), channel.WithFormats(args.In, args.Out))
	defer ch.Close()

	parse, err := mod.Exonet.Parse(chunks[0], chunks[1])
	if err != nil {
		return ch.Send(astral.Err(err))
	}
	err = mod.AddEndpoint(args.ID, parse)
	if err != nil {
		return ch.Send(astral.Err(err))
	}

	return ch.Send(&astral.Ack{})
}
