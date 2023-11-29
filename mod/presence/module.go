package presence

import (
	"context"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/admin/api"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/node/events"
	"github.com/cryptopunkscc/astrald/tasks"
	"net"
	"strconv"
)

type Module struct {
	node   node.Node
	config Config
	log    *log.Logger
	socket *net.UDPConn
	events events.Queue

	Discover *DiscoverService
	Announce *AnnounceService
}

const defaultPresencePort = 8829

func (mod *Module) Run(ctx context.Context) (err error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	if err := mod.setupSocket(); err != nil {
		return err
	}

	go func() {
		<-ctx.Done()
		mod.socket.Close()
	}()

	// inject admin command
	if adm, _ := mod.node.Modules().Find("admin").(admin.API); adm != nil {
		adm.AddCommand(ModuleName, NewAdmin(mod))
	}

	mod.log.Infov(2, "using socket %s", mod.socket.LocalAddr())

	return tasks.Group(mod.Discover, mod.Announce).Run(ctx)
}

func (mod *Module) setupSocket() (err error) {
	// resolve local address
	var localAddr *net.UDPAddr
	localAddr, err = net.ResolveUDPAddr("udp", ":"+strconv.Itoa(defaultPresencePort))
	if err != nil {
		return
	}

	// bind the udp socket
	mod.socket, err = net.ListenUDP("udp", localAddr)
	return
}

func isInterfaceEnabled(iface net.Interface) bool {
	return (iface.Flags&net.FlagUp != 0) &&
		(iface.Flags&net.FlagBroadcast != 0) &&
		(iface.Flags&net.FlagLoopback == 0)
}
