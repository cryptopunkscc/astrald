package admin

import (
	"bitbucket.org/creachadair/shell"
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/core"
	"github.com/cryptopunkscc/astrald/core/assets"
	"github.com/cryptopunkscc/astrald/debug"
	"github.com/cryptopunkscc/astrald/id"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/admin"
	"github.com/cryptopunkscc/astrald/mod/auth"
	"github.com/cryptopunkscc/astrald/mod/dir"
	"github.com/cryptopunkscc/astrald/mod/keys"
	"github.com/cryptopunkscc/astrald/sig"
	"io"
	"sync"
)

var _ core.Module = &Module{}
var _ admin.Module = &Module{}
var _ auth.Authorizer = &Module{}

const ServiceName = "admin"

type Deps struct {
	Auth auth.Module
	Dir  dir.Module
	Keys keys.Module
}

type Module struct {
	Deps
	config   Config
	node     astral.Node
	assets   assets.Assets
	admins   sig.Set[string]
	commands map[string]admin.Command
	log      *log.Logger
	mu       sync.Mutex
	ctx      context.Context
}

func (mod *Module) Run(ctx context.Context) error {
	mod.ctx = ctx

	<-ctx.Done()

	return nil
}

func (mod *Module) RouteQuery(ctx context.Context, query astral.Query, caller io.WriteCloser, hints astral.Hints) (io.WriteCloser, error) {
	if !query.Target().IsEqual(mod.node.Identity()) {
		return astral.RouteNotFound(mod)
	}

	if query.Query() != "admin" {
		return astral.RouteNotFound(mod)
	}

	// check if the caller has access to the admin panel
	if !mod.Auth.Authorize(query.Caller(), admin.ActionAccess) {
		return astral.Reject()
	}

	return astral.Accept(query, caller, mod.serve)
}

func (mod *Module) AddAdmin(identity id.Identity) error {
	return mod.admins.Add(identity.PublicKeyHex())
}

func (mod *Module) RemoveAdmin(identity id.Identity) error {
	return mod.admins.Remove(identity.PublicKeyHex())
}

func (mod *Module) hasAccess(identity id.Identity) bool {
	// Node's identity always has access to itself
	if identity.IsEqual(mod.node.Identity()) {
		return true
	}

	return mod.admins.Contains(identity.PublicKeyHex())
}

func (mod *Module) serve(conn astral.Conn) {
	defer debug.SaveLog(func(p any) {
		mod.log.Error("admin session panicked: %v", p)
	})

	defer conn.Close()

	var term = NewColorTerminal(conn, mod.log)

	for {
		term.Printf("%s@%s%s", term.UserIdentity(), mod.node.Identity(), mod.config.Prompt)

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

func (mod *Module) exec(line string, term admin.Terminal) error {
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
