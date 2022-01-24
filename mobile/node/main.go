package astralmobile

import (
	"context"
	"github.com/cryptopunkscc/astrald/app/warpdrive"
	"github.com/cryptopunkscc/astrald/app/warpdrive/service"
	"github.com/cryptopunkscc/astrald/mod/admin"
	"github.com/cryptopunkscc/astrald/mod/apphost"
	astral "github.com/cryptopunkscc/astrald/mod/apphost/client"
	"github.com/cryptopunkscc/astrald/mod/connect"
	"github.com/cryptopunkscc/astrald/mod/contacts"
	"github.com/cryptopunkscc/astrald/mod/gateway"
	"github.com/cryptopunkscc/astrald/mod/id"
	"github.com/cryptopunkscc/astrald/mod/info"
	"github.com/cryptopunkscc/astrald/node"
	"log"
	"path/filepath"
)

var identity string
var stop context.CancelFunc

func Start(astralHome string) error {
	log.Println("Staring astrald")
	astral.Instance().UseTCP = true

	// Set up app execution context
	ctx, shutdown := context.WithCancel(context.Background())

	stop = shutdown
	n, err := node.Run(
		ctx,
		astralHome,
		admin.Admin{},
		&apphost.Module{},
		connect.Connect{},
		gateway.Gateway{},
		info.Info{},
		id.Id{},
		contacts.Contacts{},
	)
	if err != nil {
		panic(err)
	}

	warpdrive.Service{
		Context: ctx,
		Api:     newApiAdapter(ctx, n),
		Config: service.Config{
			RepositoryDir:  filepath.Join(astralHome, "warpdrive"),
			RemoteResolver: true,
		},
	}.Run()

	identity = n.Identity().String()

	<-ctx.Done()

	// Run the node
	return nil
}

func Identity() string {
	return identity
}

func Stop() {
	stop()
}
