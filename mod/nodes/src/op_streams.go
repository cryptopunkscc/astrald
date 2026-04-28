package nodes

import (
	"slices"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/routing"
	"github.com/cryptopunkscc/astrald/mod/nodes"
)

type opStreamsArgs struct {
	Out string `query:"optional"`
}

// OpStreams lists all streams.
func (mod *Module) OpStreams(ctx *astral.Context, q *routing.IncomingQuery, args opStreamsArgs) (err error) {
	ch := channel.New(q.AcceptRaw(), channel.WithOutputFormat(args.Out))
	defer ch.Close()

	streams := mod.peers.streams.Clone()

	slices.SortFunc(streams, func(a, b *Link) int {
		return a.createdAt.Compare(b.createdAt)
	})

	for _, s := range streams {
		err = ch.Send(&nodes.StreamInfo{
			ID:              s.id,
			LocalIdentity:   s.LocalIdentity(),
			RemoteIdentity:  s.RemoteIdentity(),
			LocalEndpoint:   s.LocalEndpoint(),
			RemoteEndpoint:  s.RemoteEndpoint(),
			Outbound:        astral.Bool(s.outbound),
			Network:         astral.String8(s.Network()),
			HighPressure:    astral.Bool(s.PressureHigh()),
			BytesThroughput: astral.Uint64(s.Throughput()),
		})
		if err != nil {
			return ch.Send(astral.NewError(err.Error()))
		}
	}

	return ch.Send(&astral.EOS{})
}
