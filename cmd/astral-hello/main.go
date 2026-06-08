package main

import (
	"fmt"
	"io"
	"os"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/apps"
	"github.com/cryptopunkscc/astrald/lib/astrald"
	"github.com/cryptopunkscc/astrald/lib/routing"
)

// Greeting is an app-defined astral Object used to verify blueprint sync:
// after astral-hello starts, the type "astral_hello.greeting" must appear
// in objects.types on the node.
type Greeting struct {
	Recipient astral.String8
	Message   astral.String16
}

func (*Greeting) ObjectType() string                    { return "astral_hello.greeting" }
func (g Greeting) WriteTo(w io.Writer) (int64, error)   { return astral.Objectify(&g).WriteTo(w) }
func (g *Greeting) ReadFrom(r io.Reader) (int64, error) { return astral.Objectify(g).ReadFrom(r) }

func init() { _ = astral.Add(&Greeting{}) }

type App struct{}

type HelloArgs struct {
	Name string `query:"required"`
	Out  string
}

func (app *App) Hello(ctx *astral.Context, query *routing.IncomingQuery, args HelloArgs) error {
	ch := query.Accept(channel.WithOutputFormat(args.Out))
	defer ch.Close()

	return ch.Send(&Greeting{
		Recipient: astral.String8(args.Name),
		Message:   astral.String16("hello, " + args.Name + "!"),
	})
}

func main() {
	err := apps.Serve(
		astrald.NewContext(),
		routing.NewApp(&App{}),
		apps.WithBlueprintSync(),
	)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}
