package presence

import (
	"context"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/keys"
	"github.com/cryptopunkscc/astrald/mod/presence"
	"github.com/cryptopunkscc/astrald/mod/tcp"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/node/events"
	"github.com/cryptopunkscc/astrald/sig"
	"github.com/cryptopunkscc/astrald/tasks"
	"net"
	"strconv"
	"sync/atomic"
)

var _ presence.Module = &Module{}

type Module struct {
	node   node.Node
	config Config
	log    *log.Logger
	socket *net.UDPConn
	events events.Queue

	tcp  tcp.Module
	keys keys.Module

	visible   atomic.Bool
	flags     sig.Set[string] // flags attached to every ad
	flagsOnce sig.Set[string] // flags attached to the next broadcast ad only

	discover *DiscoverService
	announce *AnnounceService
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

	mod.log.Infov(2, "using socket %s", mod.socket.LocalAddr())

	return tasks.Group(mod.discover, mod.announce).Run(ctx)
}

func (mod *Module) SetVisible(b bool) error {
	if mod.visible.CompareAndSwap(!b, b) {
		if b {
			mod.flagsOnce.Add(presence.DiscoverFlag)
		}
		mod.announce.v <- b
	}

	return nil
}

func (mod *Module) myAlias() string {
	a, _ := mod.node.Tracker().GetAlias(mod.node.Identity())
	return a
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
