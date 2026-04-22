package main

import (
	"fmt"
	"os"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/apps"
	"github.com/cryptopunkscc/astrald/lib/astrald"
	"github.com/cryptopunkscc/astrald/lib/routing"
)

type App struct{}

type HelloArgs struct {
	Name string `query:"required"`
	Out  string
}

func (app *App) Hello(ctx *astral.Context, query *routing.IncomingQuery, args HelloArgs) error {
	ch := query.Accept(channel.WithOutputFormat(args.Out))
	defer ch.Close()

	return ch.Send(astral.NewString16("hello, " + args.Name + "!"))
}

func main() {
	err := apps.Serve(
		astrald.NewContext(),
		routing.NewApp(&App{}),
	)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}
