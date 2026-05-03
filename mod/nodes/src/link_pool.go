package nodes

import (
	"errors"
	"slices"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/nodes"
	"github.com/cryptopunkscc/astrald/sig"
)

type linkWatcher struct {
	match func(nodes.Link, *string) bool
	ch    chan nodes.Link
}

type LinkPool struct {
	mod      *Module
	links    sig.Map[astral.Nonce, nodes.Link]
	watchers sig.Set[*linkWatcher]
	linkers  sig.Map[string, *NodeLinker]
}

func NewLinkPool(mod *Module) *LinkPool {
	return &LinkPool{
		mod: mod,
	}
}

func (pool *LinkPool) Links() *sig.Map[astral.Nonce, nodes.Link] {
	return &pool.links
}

// SelectLinkWith returns a non-high-pressure link to id, falling back to any link if all are under pressure.
func (pool *LinkPool) SelectLinkWith(id *astral.Identity) nodes.Link {
	var fallback nodes.Link

	linksWithRemote := pool.Links().Select(func(_ astral.Nonce, a nodes.Link) bool {
		return a.RemoteIdentity().IsEqual(id)
	})

	for _, s := range linksWithRemote {
		if !s.IsHighPressure() {
			return s
		}
		if fallback == nil {
			fallback = s
		}
	}

	return fallback
}

func (pool *LinkPool) AddLink(link *Link) (tracedLink nodes.Link, err error) {
	var dir = "in"
	var netName = "unknown network"

	if link.Outbound() {
		dir = "out"
	}

	switch {
	case link.LocalEndpoint() != nil:
		netName = link.LocalEndpoint().Network()
	case link.RemoteEndpoint() != nil:
		netName = link.RemoteEndpoint().Network()
	}

	onClose := func() {
		pool.links.Delete(link.id)
		remaining := pool.links.Select(func(_ astral.Nonce, v nodes.Link) bool {
			return v.RemoteIdentity().IsEqual(link.RemoteIdentity())
		})

		pool.mod.Events.Emit(&nodes.StreamClosedEvent{
			RemoteIdentity: link.RemoteIdentity(),
			Forced:         false,
			StreamCount:    astral.Int8(len(remaining)),
		})
	}

	tl := NewTracedLink(link, onClose)

	if _, ok := pool.links.Set(link.id, tl); !ok {
		return nil, errors.New("duplicate link id")
	}

	streamsWithSameIdentity := pool.links.Select(func(_ astral.Nonce, v nodes.Link) bool {
		return v.RemoteIdentity().IsEqual(link.RemoteIdentity())
	})

	if !link.outbound {
		pool.notifyLinkWatchers(link, nil)
	}

	// fixme: LinkCreatedEvent
	pool.mod.Events.Emit(&nodes.StreamCreatedEvent{
		RemoteIdentity: link.RemoteIdentity(),
		StreamId:       link.id,
		StreamCount:    len(streamsWithSameIdentity),
	})

	pool.mod.log.Infov(1, "added %v-stream with %v (%v)", dir, link.RemoteIdentity(), netName)
	link.SetRouter(pool.mod.node)

	return tl, nil
}

type LinkResult struct {
	Stream nodes.Link
	Err    error
}

func (pool *LinkPool) subscribe(match func(nodes.Link, *string) bool) *linkWatcher {
	w := &linkWatcher{
		match: match,
		ch:    make(chan nodes.Link, 1),
	}

	pool.watchers.Add(w)
	return w
}

func (pool *LinkPool) unsubscribe(w *linkWatcher) {
	pool.watchers.Remove(w)
}

func (pool *LinkPool) notifyLinkWatchers(s nodes.Link, strategy *string) bool {
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

	match := func(s nodes.Link, strategy *string) bool {
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
		for _, tl := range pool.links.Values() {
			if match(tl, nil) {
				return sig.ArrayToChan([]LinkResult{{Stream: tl}})
			}
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
