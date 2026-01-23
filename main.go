package main

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/lib/astrald"
	apphost "github.com/cryptopunkscc/astrald/mod/apphost/client"
	"github.com/cryptopunkscc/astrald/mod/shell"
)

type Ops struct {
}

func (*Ops) Hello(ctx *astral.Context, query shell.Query) error {
	ch := query.AcceptChannel()
	defer ch.Close()

	return ch.Send(astral.NewString8("hello world"))
}

func main() {
	ctx := astrald.NewContext()

	listener, err := apphost.RegisterHandler(ctx)
	if err != nil {
		panic(err)
	}

	scope := shell.NewScope(nil)

	scope.AddStruct(&Ops{}, "")

	server := astrald.NewServer(listener, scope)

	err = server.Serve(ctx)
	if err != nil {
		panic(err)
	}
}
