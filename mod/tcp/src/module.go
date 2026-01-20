package tcp

import (
	"context"
	"fmt"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/mod/exonet"
	"github.com/cryptopunkscc/astrald/mod/tcp"
	"github.com/cryptopunkscc/astrald/mod/tree"
	"github.com/cryptopunkscc/astrald/tasks"
)

var _ tcp.Module = &Module{}

type Module struct {
	Deps
	config Config
	node   astral.Node
	log    *log.Logger
	ctx    context.Context

	configEndpoints []exonet.Endpoint

	listen       *tree.Value[*astral.Bool]
	dial         *tree.Value[*astral.Bool]
	serverCancel context.CancelFunc
}

func (mod *Module) Run(ctx *astral.Context) (err error) {
	mod.ctx = ctx

	// fixme: hacky
	mod.listen, err = mod.initTreeValue(ctx, "listen", mod.config.Listen)
	if err != nil {
		return err
	}

	mod.dial, err = mod.initTreeValue(ctx, "dial", mod.config.Dial)
	if err != nil {
		return err
	}

	// fixme: hacky
	mod.listen.OnChange(func(enabled *astral.Bool) {
		if enabled != nil {
			if *enabled {
				mod.startServer()
			} else {
				mod.stopServer()
			}
		}
	})

	if mod.ListenEnabled() {
		mod.startServer()
	}

	<-ctx.Done()
	mod.stopServer()

	return nil
}

func (mod *Module) startServer() {
	if mod.serverCancel != nil {
		return
	}

	serverCtx, cancel := context.WithCancel(mod.ctx)
	mod.serverCancel = cancel

	go func() {
		_ = tasks.Group(NewServer(mod)).Run(astral.NewContext(serverCtx))
	}()
}

func (mod *Module) stopServer() {
	if mod.serverCancel != nil {
		mod.serverCancel()
		mod.serverCancel = nil
	}
}

func (mod *Module) ListenEnabled() bool {
	if mod.listen != nil && mod.listen.Get() != nil {
		return bool(*mod.listen.Get())
	}

	return false
}

func (mod *Module) DialEnabled() bool {
	if mod.dial != nil && mod.dial.Get() != nil {
		return bool(*mod.dial.Get())
	}

	return false
}

// fixme: hacky
func (mod *Module) initTreeValue(ctx *astral.Context, name string, initial bool) (*tree.Value[*astral.Bool], error) {
	value := astral.Bool(initial)
	treeValue := tree.NewValue(&value)
	path := fmt.Sprintf(`/mod/%s/%s`, tcp.ModuleName, name)

	node, err := tree.Query(ctx, mod.Tree.Root(), path, true)
	if err != nil {
		return nil, err
	}

	err = node.Set(ctx, treeValue.Get())
	if err != nil {
		return nil, err
	}

	err = treeValue.Follow(ctx, node)
	if err != nil {
		return nil, err
	}

	return treeValue, nil
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
