package nodes

import (
	"strings"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/exonet"
	"github.com/cryptopunkscc/astrald/mod/nodes"
	"github.com/cryptopunkscc/astrald/mod/shell"
)

type opNewStreamArgs struct {
	Target   string
	Net      string `query:"optional"`
	Endpoint string `query:"optional"`
	Out      string `query:"optional"`
}

// OpNewStream initiates a new stream with the given target. If an endpoint is given, it will be used.
// If no endpoint is given, any endpoint of the given network will be used. If no network is given,
// any endpoint will be used.
func (mod *Module) OpNewStream(ctx *astral.Context, q shell.Query, args opNewStreamArgs) (err error) {
	var endpoints chan exonet.Endpoint

	target, err := mod.Dir.ResolveIdentity(args.Target)
	if err != nil {
		return q.RejectWithCode(4)
	}

	switch {
	case args.Endpoint != "":
		split := strings.SplitN(args.Endpoint, ":", 2)
		if len(split) != 2 {
			return q.RejectWithCode(2)
		}
		endpoint, err := mod.Exonet.Parse(split[0], split[1])
		if err != nil {
			return q.RejectWithCode(3)
		}

		endpoints = make(chan exonet.Endpoint, 1)
		endpoints <- endpoint
		close(endpoints)

	case args.Net != "":
		endpoints = make(chan exonet.Endpoint, 8)
		resolve, err := mod.ResolveEndpoints(ctx, target)
		if err != nil {
			mod.log.Error("resolve endpoints: %v", err)
			return q.RejectWithCode(4)
		}

		go func() {
			defer close(endpoints)
			for i := range resolve {
				if i.Network() == args.Net {
					endpoints <- i
				}
			}
		}()
	}

	ch := astral.NewChannelFmt(q.Accept(), "", args.Out)
	defer ch.Close()

	s, err := mod.peers.connectAtAny(ctx, target, endpoints)
	if err != nil {
		return ch.Write(astral.NewError(err.Error()))
	}

	return ch.Write(&nodes.StreamInfo{
		ID:             astral.Int64(s.id),
		LocalIdentity:  s.LocalIdentity(),
		RemoteIdentity: s.RemoteIdentity(),
		LocalAddr:      astral.String(s.LocalEndpoint().Address()),
		RemoteAddr:     astral.String(s.RemoteEndpoint().Address()),
		Outbound:       astral.Bool(s.outbound),
	})
}
