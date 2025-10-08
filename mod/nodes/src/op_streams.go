package nodes

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/nodes"
	"github.com/cryptopunkscc/astrald/mod/shell"
	"slices"
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
			ID:             astral.Int64(s.id),
			LocalIdentity:  s.LocalIdentity(),
			RemoteIdentity: s.RemoteIdentity(),
			LocalAddr:      astral.String(s.LocalAddr()),
			RemoteAddr:     astral.String(s.RemoteAddr()),
			Outbound:       astral.Bool(s.outbound),
		})
		if err != nil {
			return err
		}
	}

	return
}
