package nodes

import (
	"math/rand"

	"github.com/cryptopunkscc/astrald/astral"
)

type LinkPool struct {
	peers *Peers
	mod   *Module

	// todo: LinkPool is owner of streams
	// streams sig.Set[*Stream]
	// todo: linkers
}

func NewLinkPool(mod *Module, peers *Peers) *LinkPool {
	return &LinkPool{
		peers: peers,
		mod:   mod,
	}
}

type LinkFuture <-chan LinkResult

type LinkResult struct {
	Stream *Stream
	Err    error
}

func (pool *LinkPool) RetrieveLink(
	ctx *astral.Context,
	target *astral.Identity,
	endpoint *string,
	network *string,
) LinkFuture {
	result := make(chan LinkResult, 1)

	streams := pool.peers.streams.Select(func(s *Stream) bool {
		if !s.RemoteIdentity().IsEqual(target) {
			return false
		}

		if network != nil && *network != s.Network() {
			return false
		}

		if endpoint != nil && *endpoint != s.RemoteEndpoint().Address() {
			return false
		}
		return true
	})

	if len(streams) > 0 {
		rand.Shuffle(len(streams), func(i, j int) {
			streams[i], streams[j] = streams[j], streams[i]
		})

		result <- LinkResult{Stream: streams[0]}
		close(result)
		return result
	}

	// note: idk if this will be resolved like this
	endpoints, err := pool.mod.ResolveNetworkEndpoints(ctx, target, network)
	if err != nil {
		result <- LinkResult{Err: err}
		close(result)
		return result
	}

	go func() {
		defer close(result)
		// todo: subscribe to inbound connections
		// todo: handle error

		stream, err := pool.peers.connectAtAny(ctx, target, endpoints)
		// todo: wait on either stream from connectAtAny or inbound connection from this node

		result <- LinkResult{
			Stream: stream,
			Err:    err,
		}

	}()

	return result
}
