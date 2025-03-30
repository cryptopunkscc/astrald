package admin

import (
	"bitbucket.org/creachadair/shell"
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/astral/term"
	"github.com/cryptopunkscc/astrald/core"
	"github.com/cryptopunkscc/astrald/core/assets"
	"github.com/cryptopunkscc/astrald/debug"
	"github.com/cryptopunkscc/astrald/lib/query"
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

func (mod *Module) Run(ctx *astral.Context) error {
	mod.ctx = ctx

	<-ctx.Done()

	return nil
}

func (mod *Module) RouteQuery(ctx context.Context, q *astral.Query, w io.WriteCloser) (io.WriteCloser, error) {
	if !q.Target.IsEqual(mod.node.Identity()) {
		return query.RouteNotFound(mod)
	}

	if q.Query != "admin" {
		return query.RouteNotFound(mod)
	}

	// check if the caller has access to the admin panel
	if !mod.Auth.Authorize(q.Caller, admin.ActionAccess, nil) {
		return query.Reject()
	}

	return query.Accept(q, w, mod.serve)
}

func (mod *Module) AddAdmin(identity *astral.Identity) error {
	return mod.admins.Add(identity.String())
}

func (mod *Module) RemoveAdmin(identity *astral.Identity) error {
	return mod.admins.Remove(identity.String())
}

func (mod *Module) hasAccess(identity *astral.Identity) bool {
	// Node's identity always has access to itself
	if identity.IsEqual(mod.node.Identity()) {
		return true
	}

	return mod.admins.Contains(identity.String())
}

func (mod *Module) serve(conn astral.Conn) {
	defer debug.SaveLog(func(p any) {
		mod.log.Error("admin session panicked: %v", p)
	})

	defer conn.Close()

	var t = NewColorTerminal(conn, mod.log)

	for {
		t.Printf("%v@%v%v%v", t.UserIdentity(), mod.node.Identity(), mod.config.Prompt, &term.SetColor{"default"})

		line, err := t.ScanLine()
		if err != nil {
			return
		}

		if err := mod.exec(line, t); err != nil {
			t.Printf("error: %v\n", err)
		} else {
			t.Printf("ok\n")
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

func (mod *Module) String() string {
	return admin.ModuleName
}
