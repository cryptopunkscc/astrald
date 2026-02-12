package nodes

import (
	"slices"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/nodes"
	"github.com/cryptopunkscc/astrald/sig"
)

type streamWatcher struct {
	match func(*Stream) bool
	ch    chan *Stream
}

type LinkPool struct {
	peers    *Peers
	mod      *Module
	watchers sig.Set[*streamWatcher]
	linkers  sig.Map[string, *NodeLinker]
}

func NewLinkPool(mod *Module, peers *Peers) *LinkPool {
	return &LinkPool{
		peers: peers,
		mod:   mod,
	}
}

type LinkResult struct {
	Stream *Stream
	Err    error
}

func (pool *LinkPool) subscribe(match func(*Stream) bool) *streamWatcher {
	w := &streamWatcher{
		match: match,
		ch:    make(chan *Stream, 1),
	}

	pool.watchers.Add(w)
	return w
}

func (pool *LinkPool) unsubscribe(w *streamWatcher) {
	pool.watchers.Remove(w)
}

func (pool *LinkPool) notifyStreamWatchers(s *Stream) bool {
	used := false
	for _, w := range pool.watchers.Clone() {
		if !w.match(s) {
			continue
		}

		select {
		case w.ch <- s:
			used = true
		default:
		}
	}

	return used
}

func (pool *LinkPool) getOrCreateNodeLinker(target *astral.Identity) *NodeLinker {
	linker := NewNodeLinker(pool.mod, target)
	existing, ok := pool.linkers.Set(target.String(), linker)
	if !ok {
		return existing
	}

	return linker
}

func (pool *LinkPool) RetrieveLink(
	ctx *astral.Context,
	target *astral.Identity,
	opts ...RetrieveLinkOption,
) <-chan LinkResult {

	var o RetrieveLinkOptions
	for _, opt := range opts {
		opt(&o)
	}

	match := func(s *Stream) bool {
		if !s.RemoteIdentity().IsEqual(target) {
			return false
		}
		if len(o.Networks) > 0 {
			return slices.Contains(o.Networks, s.Network())
		}
		return true
	}

	if !o.ForceNew {
		streams := pool.peers.streams.Select(match)
		if len(streams) > 0 {
			return sig.ArrayToChan([]LinkResult{{Stream: streams[0]}})
		}
	}

	result := make(chan LinkResult, 1)

	go func() {
		defer close(result)
		w := pool.subscribe(match)
		defer pool.unsubscribe(w)

		linker := pool.getOrCreateNodeLinker(target)
		done := linker.Activate(ctx, o.Networks)

		select {
		case <-done:
			select {
			case s := <-w.ch:
				result <- LinkResult{Stream: s}
			default:
				result <- LinkResult{Err: nodes.ErrStreamNotProduced}
			}
		case <-ctx.Done():
			result <- LinkResult{Err: ctx.Err()}
		case s := <-w.ch:
			result <- LinkResult{Stream: s}
		}
	}()

	return result
}

// RetrieveLinkOptions controls how RetrieveLink behaves.
type RetrieveLinkOptions struct {
	ForceNew bool
	Networks []string
}

// RetrieveLinkOption is a functional option for RetrieveLink.
type RetrieveLinkOption func(*RetrieveLinkOptions)

func WithForceNew() RetrieveLinkOption {
	return func(o *RetrieveLinkOptions) {
		o.ForceNew = true
	}
}

func WithNetworks(networks ...string) RetrieveLinkOption {
	return func(o *RetrieveLinkOptions) {
		o.Networks = append(o.Networks, networks...)
	}
}
