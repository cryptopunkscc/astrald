package policy

import (
	"context"
	"github.com/cryptopunkscc/astrald/tasks"
	"time"
)

type Policy interface {
	tasks.Runner
	Name() string
}

type RunningPolicy struct {
	Policy
	startedAt time.Time
	ctx       context.Context
	cancel    context.CancelFunc
	done      chan struct{}
}

func RunPolicy(ctx context.Context, policy Policy) *RunningPolicy {
	rp := &RunningPolicy{
		Policy:    policy,
		startedAt: time.Now(),
		done:      make(chan struct{}),
	}
	rp.ctx, rp.cancel = context.WithCancel(ctx)
	go func() {
		rp.Run(rp.ctx)
		close(rp.done)
	}()
	return rp
}

func (p *RunningPolicy) Done() <-chan struct{} {
	return p.done
}

func (p *RunningPolicy) Cancel() {
	p.cancel()
}
