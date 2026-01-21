package tcp

import (
	"sync"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/mod/exonet"
	"github.com/cryptopunkscc/astrald/mod/tcp"
	"github.com/cryptopunkscc/astrald/mod/tree"
)

var _ tcp.Module = &Module{}

type Module struct {
	Deps
	config          Config
	node            astral.Node
	log             *log.Logger
	ctx             *astral.Context
	configEndpoints []exonet.Endpoint

	listen *tree.TypedBinding[*astral.Bool]
	dial   *tree.TypedBinding[*astral.Bool]

	serverMu     sync.Mutex
	serverCancel func()
}

func (mod *Module) Run(ctx *astral.Context) error {
	mod.ctx = ctx

	if v, _ := mod.listen.Value(); v != nil && *v {
		mod.startServer()
	}

	<-ctx.Done()

	return nil
}

func (mod *Module) ListenPort() int {
	return mod.config.ListenPort
}

func (mod *Module) endpoints() (list []exonet.Endpoint) {
	ips, _ := mod.IP.LocalIPs()
	for _, tip := range ips {
		e := tcp.Endpoint{
			IP:   tip,
			Port: astral.Uint16(mod.config.ListenPort),
		}

		list = append(list, &e)
	}

	return list
}

func (mod *Module) startServer() {
	mod.serverMu.Lock()
	defer mod.serverMu.Unlock()

	if mod.serverCancel != nil {
		return // already running
	}

	ctx, cancel := mod.ctx.WithCancel()
	mod.serverCancel = cancel

	go func() {
		srv := NewServer(mod)
		if err := srv.Run(ctx); err != nil {
			mod.log.Errorv(1, "server error: %v", err)
		}
	}()
}

func (mod *Module) stopServer() {
	mod.serverMu.Lock()
	defer mod.serverMu.Unlock()

	if mod.serverCancel == nil {
		return // not running
	}

	mod.serverCancel()
	mod.serverCancel = nil
}

func (mod *Module) SwitchServer(enable *astral.Bool) {
	if enable != nil && *enable {
		mod.startServer()
	} else {
		mod.stopServer()
	}
}
