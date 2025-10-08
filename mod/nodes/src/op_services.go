package nodes

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/shell"
	"strings"
)

type opServicesArgs struct {
	Nodes    string `query:"optional"`
	Services string `query:"optional"`
	Out      string `query:"optional"`
}

func (mod *Module) OpServices(ctx *astral.Context, q shell.Query, args opServicesArgs) (err error) {
	ch := astral.NewChannelFmt(q.Accept(), "", args.Out)
	defer ch.Close()

	services := mod.Services()

	if args.Nodes != "" {
		var ids []*astral.Identity
		var nodes = strings.Split(args.Nodes, ",")

		for _, node := range nodes {
			id, err := mod.Dir.ResolveIdentity(node)
			if err != nil {
				ch.Write(astral.NewError("cannot resolve node: " + node))
			}
			ids = append(ids, id)
		}

		services = services.ByNodeID(ids...)
	}

	if args.Services != "" {
		serviceNames := strings.Split(args.Services, ",")
		services = services.ByName(serviceNames...)
	}

	for _, service := range services.Find() {
		err = ch.Write(service)
		if err != nil {
			return
		}
	}
	
	return
}
