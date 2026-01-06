package nat

import (
	"github.com/cryptopunkscc/astrald/astral"
	natmod "github.com/cryptopunkscc/astrald/mod/nat"
	"github.com/cryptopunkscc/astrald/mod/services"
	"github.com/cryptopunkscc/astrald/mod/shell"
)

type opDisableArgs struct {
	Out string `query:"optional"`
}

// OpSwitch demonstrates turning off NAT's discoverable service.
// It updates the module-local current service state and publishes a corresponding ServiceChange.
func (mod *Module) OpSwitch(ctx *astral.Context, q shell.Query, args opDisableArgs) error {
	ch := astral.NewChannelFmt(q.Accept(), "", args.Out)
	defer func() { _ = ch.Close() }()

	// Update module-local state (source of truth).
	mod.serviceEnabled = !mod.serviceEnabled

	// Publish change to current subscribers.
	mod.serviceChangeFeed.Publish(services.ServiceChange{
		Enabled: astral.Bool(mod.serviceEnabled),
		Service: services.Service{
			Name:        natmod.ModuleName,
			Identity:    mod.node.Identity(),
			Composition: astral.NewBundle(),
		},
	})

	return ch.Write(&astral.EOS{})
}
