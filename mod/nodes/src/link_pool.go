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
	mod      *Module
	links    sig.Set[*Link]
	watchers sig.Set[*linkWatcher]
	linkers  sig.Map[string, *NodeLinker]
}

func NewLinkPool(mod *Module) *LinkPool {
	return &LinkPool{
		mod: mod,
	}
}

func (pool *LinkPool) Links() *sig.Set[*Link] {
	return &pool.links
}

// SelectLinkWith returns a non-high-pressure link to id, falling back to any link if all are under pressure.
func (pool *LinkPool) SelectLinkWith(id *astral.Identity) *Link {
	var fallback *Link
	for _, s := range pool.links.Clone() {
		if !s.RemoteIdentity().IsEqual(id) {
			continue
		}
		if !s.IsHighPressure() {
			return s
		}
		if fallback == nil {
			fallback = s
		}
	}
	return fallback
}

func (pool *LinkPool) AddLink(link *Link) (*TracedLink, error) {
	if err := pool.links.Add(link); err != nil {
		return nil, err
	}

	streamsWithSameIdentity := pool.links.Select(func(v *Link) bool {
		return v.RemoteIdentity().IsEqual(link.RemoteIdentity())
	})

	if !link.outbound {
		pool.notifyLinkWatchers(link, nil)
	}

	pool.mod.Events.Emit(&nodes.StreamCreatedEvent{
		RemoteIdentity: link.RemoteIdentity(),
		StreamId:       link.id,
		StreamCount:    len(streamsWithSameIdentity),
	})

	tl := NewTracedLink(link, func() {
		pool.links.Remove(link)

		remaining := pool.links.Select(func(v *Link) bool {
			return v.RemoteIdentity().IsEqual(link.RemoteIdentity())
		})

		pool.mod.Events.Emit(&nodes.StreamClosedEvent{
			RemoteIdentity: link.RemoteIdentity(),
			Forced:         false,
			StreamCount:    astral.Int8(len(remaining)),
		})
	})

	return tl, nil
}

type LinkResult struct {
	Stream *Link
	Err    error
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
		streams := pool.links.Select(func(s *Link) bool { return match(s, nil) })
		if len(streams) > 0 {
			return sig.ArrayToChan([]LinkResult{{Stream: streams[0]}})
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
