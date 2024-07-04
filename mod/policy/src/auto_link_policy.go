package policy

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/net"
	"strings"
	"sync"
)

var _ Policy = &AutoLinkPolicy{}

type AutoLinkPolicy struct {
	*Module
	active map[string]chan struct{}
	mu     sync.Mutex
}

func NewAutoLinkPolicy(module *Module) *AutoLinkPolicy {
	return &AutoLinkPolicy{
		Module: module,
		active: map[string]chan struct{}{},
	}
}

func (policy *AutoLinkPolicy) Run(ctx context.Context) error {
	policy.node.Router().AddRoute(policy.node.Identity(), id.Anyone, policy, 0)
	defer policy.node.Router().RemoveRoute(policy.node.Identity(), id.Anyone, policy)

	<-ctx.Done()
	return nil
}

func (policy *AutoLinkPolicy) RouteQuery(ctx context.Context, query net.Query, caller net.SecureWriteCloser, hints net.Hints) (targetWriter net.SecureWriteCloser, err error) {
	// check if autolinker is already running for the target
	if wait, locked := policy.lock(query.Target()); locked {
		// if we locked successfully, unlock at when we're done
		policy.log.Logv(2, "autolinking with %v...", query.Target())
		defer policy.unlock(query.Target())
	} else {
		// if target is already locked, wait for the locker to finish and reroute
		policy.log.Logv(2, "autolinking already in progress", query.Target())
		select {
		case <-wait:
			// try to reroute after the other linker is done
			return policy.node.Router().RouteQuery(ctx, query, caller, hints.SetReroute())
		case <-ctx.Done():
			return nil, net.ErrTimeout
		}
	}

	// don't open a new link with the target if we have one already
	if policy.node.Network().Links().ByRemoteIdentity(query.Target()).Count() > 0 {
		return net.RouteNotFound(policy)
	}

	_, err = policy.node.Network().Link(ctx, query.Target())
	if err != nil {
		switch {
		case errors.Is(err, context.Canceled):
			return nil, net.ErrAborted

		case errors.Is(err, context.DeadlineExceeded):
			return nil, net.ErrTimeout

		case strings.Contains(err.Error(), "no endpoints provided"):
			return net.RouteNotFound(policy)
		}

		return net.RouteNotFound(policy, err)
	}

	// reroute the query
	return policy.node.Router().RouteQuery(ctx, query, caller, hints.SetReroute())
}

func (policy *AutoLinkPolicy) Name() string {
	return "auto_link"
}

func (policy *AutoLinkPolicy) lock(id id.Identity) (chan struct{}, bool) {
	policy.mu.Lock()
	defer policy.mu.Unlock()

	var hex = id.PublicKeyHex()

	if active, found := policy.active[hex]; found {
		return active, false
	}

	policy.active[hex] = make(chan struct{})

	return policy.active[hex], true
}

func (policy *AutoLinkPolicy) unlock(id id.Identity) {
	policy.mu.Lock()
	defer policy.mu.Unlock()

	var hex = id.PublicKeyHex()

	if active, found := policy.active[hex]; found {
		delete(policy.active, hex)
		close(active)
	}
}
