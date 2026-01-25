package nodes

import (
	"slices"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/ops"
	"github.com/cryptopunkscc/astrald/mod/nodes"
)

type opStreamsArgs struct {
	Out string `query:"optional"`
}

// OpStreams lists all streams.
func (mod *Module) OpStreams(ctx *astral.Context, q *ops.Query, args opStreamsArgs) (err error) {
	ch := channel.New(q.Accept(), channel.WithOutputFormat(args.Out))
	defer ch.Close()

	streams := mod.peers.streams.Clone()

	slices.SortFunc(streams, func(a, b *Stream) int {
		return a.createdAt.Compare(b.createdAt)
	})

	for _, s := range streams {
		err = ch.Send(&nodes.StreamInfo{
			ID:             s.id,
			LocalIdentity:  s.LocalIdentity(),
			RemoteIdentity: s.RemoteIdentity(),
			LocalEndpoint:  s.LocalEndpoint(),
			RemoteEndpoint: s.RemoteEndpoint(),
			Outbound:       astral.Bool(s.outbound),
			Network:        astral.String8(s.Network()),
		})
		if err != nil {
			return ch.Send(astral.NewError(err.Error()))
		}
	}

	return ch.Send(&astral.EOS{})
}
