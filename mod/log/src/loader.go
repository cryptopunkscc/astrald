package log

import (
	"os"

	"github.com/cryptopunkscc/astrald/astral"
	alog "github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/core"
	"github.com/cryptopunkscc/astrald/core/assets"
	modlog "github.com/cryptopunkscc/astrald/mod/log"
)

type Loader struct{}

func (Loader) Load(node astral.Node, assets assets.Assets, log *alog.Logger) (core.Module, error) {
	var err error
	var mod = &Module{
		node:   node,
		config: defaultConfig,
		log:    log,
		assets: assets,
	}

	_ = assets.LoadYAML(modlog.ModuleName, &mod.config)

	err = mod.ops.AddStruct(mod, "Op")
	if err != nil {
		return nil, err
	}

	alog.DefaultViewer.Set(
		astral.Identity{}.ObjectType(),
		func(object astral.Object) astral.Object {
			identity := object.(*astral.Identity)

			return modlog.IdentityView{
				Identity:  identity,
				Highlight: identity.IsEqual(node.Identity()),
			}
		},
	)

	// add default output to the output list
	p := alog.NewPrinter(os.Stdout)
	p.Filter = mod.LogEntryFilter
	mod.outputs.Add(p)

	// add a log file to the output list
	logFile, err := CreateLogFile()
	if err != nil {
		log.Error("cannot create log file: %v", err)
	} else {
		mod.outputs.Add(logFile)
	}

	// switch logger output to the module
	log.SetOutput(mod)

	return mod, err
}

func init() {
	if err := core.RegisterModule(modlog.ModuleName, Loader{}); err != nil {
		panic(err)
	}
}
