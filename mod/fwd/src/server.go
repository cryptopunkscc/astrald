package fwd

import (
	"context"
	"fmt"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/tasks"
	"time"
)

type Server interface {
	tasks.Runner
	fmt.Stringer
	Target() astral.Router
}

type ServerRunner struct {
	Server
	startedAt time.Time
	ctx       *astral.Context
	cancel    context.CancelFunc
	err       error
	done      chan struct{}
}

func NewServerRunner(ctx *astral.Context, s Server) *ServerRunner {
	ctx, cancel := ctx.WithCancel()

	return &ServerRunner{
		Server:    s,
		startedAt: time.Now(),
		ctx:       ctx,
		cancel:    cancel,
	}
}

func (srv *ServerRunner) Run(ctx *astral.Context) error {
	srv.done = make(chan struct{})
	defer close(srv.done)
	srv.ctx = ctx
	srv.err = srv.Server.Run(ctx)
	return srv.err
}

func (srv *ServerRunner) Done() <-chan struct{} {
	return srv.done
}

func (srv *ServerRunner) String() string {
	return srv.Server.String()
}

func (srv *ServerRunner) Stop() error {
	srv.cancel()
	return nil
}
