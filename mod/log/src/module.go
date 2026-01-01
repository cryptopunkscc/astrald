package log

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/mod/dir"
	modlog "github.com/cryptopunkscc/astrald/mod/log"
	"github.com/cryptopunkscc/astrald/mod/shell"
	"github.com/cryptopunkscc/astrald/resources"
	"github.com/cryptopunkscc/astrald/sig"
)

type Deps struct {
	Dir dir.Module
}

type Module struct {
	Deps
	config      Config
	node        astral.Node
	log         *log.Logger
	assets      resources.Resources
	outputs     sig.Set[log.Output]
	logFilePath string
	ops         shell.Scope
}

func (mod *Module) LogEntry(entry *log.Entry) {
	for _, output := range mod.outputs.Clone() {
		output.LogEntry(entry)
	}
}

func (mod *Module) LogEntryFilter(entry *log.Entry) bool {
	if entry.Level > mod.config.Level {
		return false
	}

	return true
}

func (mod *Module) Run(ctx *astral.Context) error {
	<-ctx.Done()
	return nil
}

func (mod *Module) Scope() *shell.Scope {
	return &mod.ops
}

func (mod *Module) String() string {
	return modlog.ModuleName
}
