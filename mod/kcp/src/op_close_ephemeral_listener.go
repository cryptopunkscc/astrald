package kcp

import (
	"strings"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/exonet"
	"github.com/cryptopunkscc/astrald/mod/shell"
)

type opCloseEphemeralListenerArgs struct {
	Endpoint string
	In       string `query:"optional"`
	Out      string `query:"optional"`
}

func (mod *Module) OpCloseEphemeralListener(ctx *astral.Context, q shell.Query, args opCloseEphemeralListenerArgs) (err error) {
	chunks := strings.SplitN(args.Endpoint, ":", 2)
	if len(chunks) != 2 {
		return q.RejectWithCode(2)
	}

	_, err = mod.Exonet.Parse(chunks[0], chunks[1])
	if err != nil {
		return q.RejectWithCode(2)
	}

	mod.ephemeralListeners.Each(func(k string, v exonet.EphemeralListener) error {
		mod.log.Log(`%v`, k)

		return nil
	})

	listener, ok := mod.ephemeralListeners.Get(chunks[1])
	if !ok {
		return q.RejectWithCode(4)
	}

	ch := astral.NewChannelFmt(q.Accept(), args.In, args.Out)
	defer ch.Close()

	err = listener.Close()
	if err != nil {
		return ch.Write(astral.NewError(err.Error()))
	}

	return ch.Write(&astral.Ack{})
}
