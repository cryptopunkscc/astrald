package auth

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/mod/auth"
	"github.com/cryptopunkscc/astrald/resources"
	"github.com/cryptopunkscc/astrald/sig"
)

type Module struct {
	Deps
	config   Config
	node     astral.Node
	log      *log.Logger
	assets   resources.Resources
	db       *DB
	handlers sig.Map[string, []auth.Handler]
}

func (mod *Module) Run(ctx *astral.Context) error {
	return nil
}

// Add registers handlers for the given action ObjectType string.
func (mod *Module) Add(actionType string, handlers ...auth.Handler) {
	mod.handlers.Set(actionType, append(mod.get(actionType), handlers...))
}

func (mod *Module) get(actionType string) []auth.Handler {
	h, _ := mod.handlers.Get(actionType)
	return h
}
