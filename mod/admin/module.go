package admin

import (
	"bitbucket.org/creachadair/shell"
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/debug"
	"github.com/cryptopunkscc/astrald/log"
	. "github.com/cryptopunkscc/astrald/mod/admin/api"
	"github.com/cryptopunkscc/astrald/mod/router/api"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/node/assets"
	"github.com/cryptopunkscc/astrald/node/modules"
	"sync"
)

var _ modules.Module = &Module{}
var _ API = &Module{}

const ServiceName = "admin"

type Module struct {
	config   Config
	node     node.Node
	assets   assets.Store
	commands map[string]Command
	log      *log.Logger
	mu       sync.Mutex
	ctx      context.Context
	router   router.API
}

func (mod *Module) Prepare(ctx context.Context) error {
	mod.router, _ = router.Load(mod.node)

	return nil
}

func (mod *Module) Run(ctx context.Context) error {
	mod.ctx = ctx

	err := mod.node.LocalRouter().AddRoute(ServiceName, mod)
	if err != nil {
		return err
	}
	defer mod.node.LocalRouter().RemoveRoute(ServiceName)

	<-ctx.Done()

	return nil
}

func (mod *Module) RouteQuery(ctx context.Context, query net.Query, caller net.SecureWriteCloser, hints net.Hints) (net.SecureWriteCloser, error) {
	// check if the caller has access to the admin panel
	if !mod.hasAccess(caller.Identity()) {
		mod.log.Errorv(1, "denied access to %v", caller.Identity())
		return net.Reject()
	}

	mod.log.Info("%v has accessed the admin panel", caller.Identity())
	return net.Accept(query, caller, mod.serve)
}

func (mod *Module) hasAccess(identity id.Identity) bool {
	// Node's identity always has access to itself
	if identity.IsEqual(mod.node.Identity()) {
		return true
	}

	// Check config file admins
	for _, name := range mod.config.Admins {
		admin, err := mod.node.Resolver().Resolve(name)
		if err != nil {
			continue
		}

		if identity.IsEqual(admin) {
			return true
		}
	}

	return false
}

func (mod *Module) serve(conn net.SecureConn) {
	defer debug.SaveLog(func(p any) {
		mod.log.Error("admin session panicked: %v", p)
	})

	defer conn.Close()

	var term = NewColorTerminal(conn, mod.log)

	for {
		term.Printf("%s@%s%s", conn.RemoteIdentity(), mod.node.Identity(), mod.config.Prompt)

		line, err := term.ScanLine()
		if err != nil {
			return
		}

		if err := mod.exec(line, term); err != nil {
			term.Printf("error: %v\n", err)
		} else {
			term.Printf("ok\n")
		}
	}
}

func (mod *Module) exec(line string, term Terminal) error {
	args, valid := shell.Split(line)
	if len(args) == 0 {
		return nil
	}
	if !valid {
		return errors.New("unclosed quotes")
	}

	if cmd, found := mod.commands[args[0]]; found {
		return cmd.Exec(term, args)
	} else {
		return errors.New("command not found")
	}
}
