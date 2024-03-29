package dir

import (
	"context"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/lib/desc"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/dir"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/resources"
	"github.com/cryptopunkscc/astrald/sig"
	"github.com/cryptopunkscc/astrald/tasks"
)

var _ dir.Module = &Module{}

type Module struct {
	config Config
	node   node.Node
	log    *log.Logger
	assets resources.Resources

	describers sig.Set[dir.Describer]
}

func (mod *Module) Run(ctx context.Context) error {
	return tasks.Group(
		&Service{Module: mod},
	).Run(ctx)
}

func (mod *Module) Describe(ctx context.Context, identity id.Identity, opts *desc.Opts) []*desc.Desc {
	var list []desc.Describer[id.Identity]

	for _, d := range mod.describers.Clone() {
		list = append(list, d)
	}

	return desc.Collect(ctx, identity, opts, list...)
}

func (mod *Module) AddDescriber(describer dir.Describer) error {
	return mod.describers.Add(describer)
}

func (mod *Module) RemoveDescriber(describer dir.Describer) error {
	return mod.describers.Remove(describer)
}
