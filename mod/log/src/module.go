package log

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/lib/ops"
	"github.com/cryptopunkscc/astrald/mod/dir"
	modlog "github.com/cryptopunkscc/astrald/mod/log"
	"github.com/cryptopunkscc/astrald/mod/tree"
	"github.com/cryptopunkscc/astrald/resources"
	"github.com/cryptopunkscc/astrald/sig"
)

type Deps struct {
	Dir  dir.Module
	Tree tree.Module
}

type Module struct {
	Deps
	config      Config
	node        astral.Node
	log         *log.Logger
	assets      resources.Resources
	outputs     sig.Set[log.Output]
	logFilePath string
	ops         ops.Set
}

func (mod *Module) LogEntry(entry *log.Entry) {
	for _, output := range mod.outputs.Clone() {
		output.LogEntry(entry)
	}
}

func (mod *Module) LogEntryFilter(entry *log.Entry) bool {
	lvl := (*uint8)(mod.config.Level.Get())
	if lvl == nil {
		return entry.Level <= DefaultLogLevel
	}
	return entry.Level <= *lvl
}

func (mod *Module) Run(ctx *astral.Context) error {
	<-ctx.Done()
	return nil
}

func (mod *Module) GetOpSet() *ops.Set {
	return &mod.ops
}

func (mod *Module) String() string {
	return modlog.ModuleName
}
