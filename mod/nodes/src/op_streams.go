package nodes

import (
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

	streams := mod.linkPool.Links().Values()
	for _, s := range streams {
		err = ch.Send(&nodes.StreamInfo{
			ID:              s.ID(),
			LocalIdentity:   s.LocalIdentity(),
			RemoteIdentity:  s.RemoteIdentity(),
			LocalEndpoint:   s.LocalEndpoint(),
			RemoteEndpoint:  s.RemoteEndpoint(),
			Outbound:        astral.Bool(s.Outbound()),
			Network:         astral.String8(s.Network()),
			HighPressure:    astral.Bool(s.IsHighPressure()),
			BytesThroughput: astral.Uint64(s.Throughput()),
		})
		if err != nil {
			return ch.Send(astral.Err(err))
		}
	}

	return ch.Send(&astral.EOS{})
}
