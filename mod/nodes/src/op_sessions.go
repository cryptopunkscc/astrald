package nodes

import (
	"slices"
	"time"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/routing"
	"github.com/cryptopunkscc/astrald/mod/nodes"
)

type opSessionsArgs struct {
	Out string `query:"optional"`
}

// OpSessions lists all active sessions.
func (mod *Module) OpSessions(ctx *astral.Context, q *routing.IncomingQuery, args opSessionsArgs) (err error) {
	ch := channel.New(q.AcceptRaw(), channel.WithOutputFormat(args.Out))
	defer ch.Close()

	sessions := mod.peers.sessions.Values()

	slices.SortFunc(sessions, func(a, b *session) int {
		return a.createdAt.Compare(b.createdAt)
	})

	for _, s := range sessions {
		if !s.IsOpen() {
			continue
		}

		s.mu.Lock()
		stream := s.stream
		s.mu.Unlock()

		err = ch.Send(&nodes.SessionInfo{
			ID:             s.Nonce,
			StreamID:       stream.id,
			RemoteIdentity: s.RemoteIdentity,
			Outbound:       astral.Bool(s.Outbound),
			Query:          astral.String16(s.Query),
			Bytes:          astral.Uint64(s.bytes.Load()),
			Age:            astral.Duration(time.Since(s.createdAt)),
			CanMigrate:     astral.Bool(s.CanMigrate()),
		})
		if err != nil {
			return ch.Send(astral.NewError(err.Error()))
		}
	}

	return ch.Send(&astral.EOS{})
}
