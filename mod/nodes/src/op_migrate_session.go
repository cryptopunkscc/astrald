package nodes

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/shell"
)

type opMigrateSessionArgs struct {
	Nonce astral.Nonce // session that should be migrated
	// TargetStreamId specifies the ID of the target stream for migration.
	TargetStreamId astral.Uint64
	//
	Out string `query:"optional"`
}

func (mod *Module) OpMigrateSession(ctx *astral.Context, q shell.Query, args opMigrateSessionArgs) (err error) {
	ch := astral.NewChannelFmt(q.Accept(), "", args.Out)
	defer ch.Close()

	session, ok := mod.peers.sessions.Get(args.Nonce)
	if !ok {
		return ch.Write(astral.NewError("session not found"))
	}

	if session.Outbound {
		return ch.Write(astral.NewError("session is outbound"))
	}

	if session.state != stateOpen {
		return ch.Write(astral.NewError("session is not open"))
	}

	targetStream := mod.peers.streams.Select(func(a *Stream) bool {
		if a.id == int(args.TargetStreamId) {
			return true

		}
		return false
	})

	if len(targetStream) == 0 {
		return ch.Write(astral.NewError("target stream not found"))
	}

	if session.stream.id == targetStream[0].id {
		return ch.Write(astral.NewError("target stream is the same as the source stream"))
	}

	return
}
