package fwd

import (
	"context"
	"errors"
	"fmt"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/admin/api"
	"github.com/cryptopunkscc/astrald/mod/tcp/api"
	"github.com/cryptopunkscc/astrald/mod/tor/api"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/node/assets"
	"github.com/cryptopunkscc/astrald/node/modules"
	"strings"
	"sync"
)

type Module struct {
	node    modules.Node
	assets  assets.Store
	config  Config
	log     *log.Logger
	ctx     context.Context
	servers map[*ServerRunner]struct{}
	mu      sync.Mutex
	tcp     tcp.API
	tor     tor.API
}

func (mod *Module) Run(ctx context.Context) error {
	mod.ctx = ctx

	// inject admin command
	if adm, _ := mod.node.Modules().Find("admin").(admin.API); adm != nil {
		adm.AddCommand(ModuleName, NewAdmin(mod))
	}

	mod.tcp, _ = mod.node.Modules().Find("tcp").(tcp.API)
	mod.tor, _ = mod.node.Modules().Find("tor").(tor.API)

	for server, target := range mod.config.Forwards {
		err := mod.CreateForward(server, target)
		if err != nil {
			mod.log.Errorv(1, "error creating %v -> %v: %v",
				server,
				target,
				err,
			)
		}
	}

	<-ctx.Done()
	mod.waitForServers()

	return nil
}

func (mod *Module) CreateForward(server, target string) error {
	t, err := mod.parseTarget(target)
	if err != nil {
		return fmt.Errorf("cannot parse target: %w", err)
	}

	s, err := mod.createServer(server, t)
	if err != nil {
		return fmt.Errorf("cannot create server: %w", err)
	}

	return mod.runServer(s)
}

func (mod *Module) Servers() []*ServerRunner {
	mod.mu.Lock()
	defer mod.mu.Unlock()

	var list = make([]*ServerRunner, 0)

	for server := range mod.servers {
		list = append(list, server)
	}

	return list
}

func (mod *Module) waitForServers() {
	mod.mu.Lock()
	defer mod.mu.Unlock()

	for server := range mod.servers {
		<-server.Done()
	}
}

func (mod *Module) runServer(s *ServerRunner) error {
	mod.mu.Lock()
	defer mod.mu.Unlock()

	mod.servers[s] = struct{}{}

	go func() {
		defer s.Stop()

		err := s.Run(s.ctx)

		if err != nil {
			mod.log.Errorv(1, "%v ended with error: %v", s, err)
		}

		mod.mu.Lock()
		defer mod.mu.Unlock()

		delete(mod.servers, s)
	}()

	return nil
}

func (mod *Module) parseTarget(uri string) (net.Router, error) {
	var err error
	var idx = strings.Index(uri, "://")
	if idx == -1 {
		return nil, errors.New("missing protocol")
	}

	var proto = uri[:idx]
	uri = uri[idx+3:]

	switch proto {
	case "tcp":
		return NewTCPTarget(uri, mod.node.Identity())

	case "astral":
		var caller = mod.node.Identity()
		var target = mod.node.Identity()

		if idx := strings.Index(uri, "@"); idx != -1 {
			callerName := uri[:idx]
			uri = uri[idx+1:]

			caller, err = mod.node.Resolver().Resolve(callerName)
			if err != nil {
				return nil, err
			}

			// fetch private key if we're calling as non-node identity
			if !caller.IsEqual(mod.node.Identity()) {
				keystore, err := mod.assets.KeyStore()
				if err != nil {
					return nil, err
				}

				caller, err = keystore.Find(caller)
				if err != nil {
					return nil, err
				}
			}
		}

		if idx := strings.Index(uri, ":"); idx != -1 {
			name := uri[:idx]
			uri = uri[idx+1:]

			target, err = mod.node.Resolver().Resolve(name)
			if err != nil {
				return nil, err
			}
		}

		if len(uri) == 0 {
			return nil, errors.New("missing query")
		}

		var query = net.NewQuery(caller, target, uri)

		var label = fmt.Sprintf("%s@%s:%s",
			mod.node.Resolver().DisplayName(caller),
			mod.node.Resolver().DisplayName(target),
			uri,
		)

		return NewAstralTarget(query, mod.node.Router(), label)

	case "tor":
		if mod.tor == nil {
			return nil, errors.New("tor driver not found")
		}

		return NewTorTarget(mod.tor, uri, mod.node.Identity())

	default:
		return nil, errors.New("unsupported protocol")
	}
}

func (mod *Module) createServer(uri string, target net.Router) (*ServerRunner, error) {
	var idx = strings.Index(uri, "://")
	if idx == -1 {
		return nil, errors.New("missing protocol")
	}

	var proto = uri[:idx]
	uri = uri[idx+3:]

	switch proto {
	case "tcp":
		tcpServer, err := NewTCPServer(mod, uri, target)
		if err != nil {
			return nil, err
		}

		return NewServerRunner(mod.ctx, tcpServer), nil

	case "astral":
		astralServer, err := NewAstralServer(mod, uri, target)
		if err != nil {
			return nil, err
		}

		return NewServerRunner(mod.ctx, astralServer), nil

	default:
		return nil, errors.New("unsupported protocol")
	}
}
