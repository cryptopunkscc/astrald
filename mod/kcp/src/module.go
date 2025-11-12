package kcp

import (
	"fmt"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/core/assets"
	"github.com/cryptopunkscc/astrald/mod/exonet"
	"github.com/cryptopunkscc/astrald/mod/kcp"
	"github.com/cryptopunkscc/astrald/mod/shell"
	"github.com/cryptopunkscc/astrald/sig"
	"github.com/cryptopunkscc/astrald/tasks"
)

// Module represents the KCP module and implements the exonet.Dialer interface.
type Module struct {
	Deps
	config Config
	node   astral.Node
	assets assets.Assets
	log    *log.Logger
	ctx    *astral.Context
	ops    shell.Scope

	configEndpoints       []exonet.Endpoint
	ephemeralListeners    sig.Map[string, exonet.EphemeralListener]
	ephemeralPortMappings sig.Map[string, astral.Uint16]
}

func (mod *Module) Run(ctx *astral.Context) error {
	mod.ctx = ctx

	err := tasks.Group(NewServer(mod, mod.config.ListenPort, mod.Nodes.Accept)).Run(ctx)
	if err != nil {
		return err
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
		e := kcp.Endpoint{
			IP:   tip,
			Port: astral.Uint16(mod.config.ListenPort),
		}

		list = append(list, &e)
	}

	for endpointStr, _ := range mod.ephemeralListeners.Clone() {
		e, err := kcp.ParseEndpoint(endpointStr)
		if err != nil {
			continue
		}

		list = append(list, e)
	}

	return list
}

// CreateEphemeralListener creates an ephemeral KCP endpoint which will start server that listens on the specified local endpoint.
// and push the endpoint to the ephemeralListeners set.
func (mod *Module) CreateEphemeralListener(local kcp.Endpoint) (err error) {
	kcpServer := NewServer(mod, local.UDPAddr().Port, mod.Nodes.Accept)

	// Run server asynchronously (non-blocking)
	go func() {
		if err := tasks.Group(kcpServer).Run(mod.ctx); err != nil {
			mod.log.Error("ephemeral listener error: %v", err)
		}

		// cleanup after it exits
		mod.ephemeralListeners.Delete(local.Address())
		<-mod.ctx.Done()
	}()

	_, ok := mod.ephemeralListeners.Set(local.Address(), kcpServer)
	if !ok {
		// NOTE: such server should never start in first place, if we could not add it to map
		kcpServer.Close()
	}

	return nil
}

func (mod *Module) SetEndpointLocalSocket(e kcp.Endpoint, localSocket astral.Uint16) error {
	_, ok := mod.ephemeralPortMappings.Get(e.Address())
	if ok {
		// mapping already exists (?) for now treat as error
		return fmt.Errorf("mapping for endpoint %s already exists", e.Address())
	}

	mod.ephemeralPortMappings.Set(e.Address(), localSocket)
	return nil
}

func (mod *Module) Scope() *shell.Scope {
	return &mod.ops
}

func (mod *Module) String() string {
	return kcp.ModuleName
}
