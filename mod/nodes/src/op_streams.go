package nodes

import (
	"slices"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/nodes"
	"github.com/cryptopunkscc/astrald/mod/shell"
)

type opStreamsArgs struct {
	Out string `query:"optional"`
}

// OpStreams lists all streams.
func (mod *Module) OpStreams(ctx *astral.Context, q shell.Query, args opStreamsArgs) (err error) {
	ch := astral.NewChannelFmt(q.Accept(), "", args.Out)
	defer ch.Close()

	streams := mod.peers.streams.Clone()

	slices.SortFunc(streams, func(a, b *Stream) int {
		return a.createdAt.Compare(b.createdAt)
	})

	for _, s := range streams {
		err = ch.Write(&nodes.StreamInfo{
			ID:             s.id,
			LocalIdentity:  s.LocalIdentity(),
			RemoteIdentity: s.RemoteIdentity(),
			LocalEndpoint:  s.LocalEndpoint(),
			RemoteEndpoint: s.RemoteEndpoint(),
			Outbound:       astral.Bool(s.outbound),
			Network:        astral.String8(s.Network()),
		})
		if err != nil {
			return ch.Write(astral.NewError(err.Error()))
		}
	}

	return ch.Write(&astral.EOS{})
}
