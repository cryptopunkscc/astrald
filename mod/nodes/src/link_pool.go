package nodes

import (
	"slices"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/nodes"
	"github.com/cryptopunkscc/astrald/sig"
)

type linkWatcher struct {
	match func(*Link, *string) bool
	ch    chan *Link
}

type LinkPool struct {
	peers    *Peers
	mod      *Module
	watchers sig.Set[*linkWatcher]
	linkers  sig.Map[string, *NodeLinker]
}

func NewLinkPool(mod *Module, peers *Peers) *LinkPool {
	return &LinkPool{
		peers: peers,
		mod:   mod,
	}
}

type LinkResult struct {
	Link *Link
	Err  error
}

func (pool *LinkPool) subscribe(match func(*Link, *string) bool) *linkWatcher {
	w := &linkWatcher{
		match: match,
		ch:    make(chan *Link, 1),
	}

	pool.watchers.Add(w)
	return w
}

func (pool *LinkPool) unsubscribe(w *linkWatcher) {
	pool.watchers.Remove(w)
}

func (pool *LinkPool) notifyLinkWatchers(s *Link, strategy *string) bool {
	used := false
	for _, w := range pool.watchers.Clone() {
		if !w.match(s, strategy) {
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
	key := target.String()

	if linker, ok := pool.linkers.Get(key); ok {
		return linker
	}

	linker := NewNodeLinker(pool.mod, target)
	if existing, ok := pool.linkers.Set(key, linker); !ok {
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

	match := func(s *Link, strategy *string) bool {
		if !s.RemoteIdentity().IsEqual(target) {
			return false
		}
		if strategy != nil && len(o.Strategies) > 0 {
			return slices.Contains(o.Strategies, *strategy)
		}

		if len(o.Networks) > 0 {
			return slices.Contains(o.Networks, s.Network())
		}
		return true
	}

	if !o.ForceNew {
		links := pool.peers.links.Select(func(s *Link) bool { return match(s, nil) })
		if len(links) > 0 {
			return sig.ArrayToChan([]LinkResult{{Link: links[0]}})
		}
	}

	result := make(chan LinkResult, 1)

	go func() {
		defer close(result)

		strategyCtx, cancelStrategies := ctx.WithCancel()
		defer cancelStrategies()

		w := pool.subscribe(match)
		defer pool.unsubscribe(w)

		linker := pool.getOrCreateNodeLinker(target)
		done := linker.Activate(strategyCtx, o.Strategies)

		select {
		case <-done:
			select {
			case s := <-w.ch:
				result <- LinkResult{Link: s}
			default:
				result <- LinkResult{Err: nodes.ErrLinkNotProduced}
			}
		case <-ctx.Done():
			result <- LinkResult{Err: ctx.Err()}
		case s := <-w.ch:
			result <- LinkResult{Link: s}
		}
	}()

	return result
}

// RetrieveLinkOptions controls how RetrieveLink behaves.
type RetrieveLinkOptions struct {
	ForceNew   bool
	Strategies []string
	Networks   []string
}

// RetrieveLinkOption is a functional option for RetrieveLink.
type RetrieveLinkOption func(*RetrieveLinkOptions)

func WithForceNew() RetrieveLinkOption {
	return func(o *RetrieveLinkOptions) {
		o.ForceNew = true
	}
}

func WithStrategies(strategies ...string) RetrieveLinkOption {
	return func(o *RetrieveLinkOptions) {
		o.Strategies = append(o.Strategies, strategies...)
	}
}

func WithNetworks(networks ...string) RetrieveLinkOption {
	return func(o *RetrieveLinkOptions) {
		o.Networks = append(o.Networks, networks...)
	}
}
