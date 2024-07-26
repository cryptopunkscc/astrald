package presence

import (
	"context"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/events"
	"github.com/cryptopunkscc/astrald/lib/routers"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/admin"
	"github.com/cryptopunkscc/astrald/mod/auth"
	"github.com/cryptopunkscc/astrald/mod/dir"
	"github.com/cryptopunkscc/astrald/mod/keys"
	"github.com/cryptopunkscc/astrald/mod/nodes"
	"github.com/cryptopunkscc/astrald/mod/presence"
	"github.com/cryptopunkscc/astrald/mod/tcp"
	"github.com/cryptopunkscc/astrald/sig"
	"github.com/cryptopunkscc/astrald/tasks"
	"net"
	"strconv"
	"sync/atomic"
	"time"
)

var _ presence.Module = &Module{}

type Deps struct {
	Admin admin.Module
	Auth  auth.Module
	Dir   dir.Module
	Keys  keys.Module
	Nodes nodes.Module
	TCP   tcp.Module
}

type Module struct {
	Deps
	*routers.PathRouter
	node   astral.Node
	config Config
	log    *log.Logger
	socket *net.UDPConn
	events events.Queue

	visible    atomic.Bool
	outFilters sig.Set[presence.AdOutHook]

	discover *DiscoverService
	announce *AnnounceService
}

const defaultPresencePort = 8829

func (mod *Module) Run(ctx context.Context) (err error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	if mod.config.Discoverable {
		go func() {
			// wait for services to start
			time.Sleep(time.Second)
			mod.SetVisible(true)
		}()
	}

	if err := mod.setupSocket(); err != nil {
		return err
	}

	go func() {
		<-ctx.Done()
		mod.socket.Close()
	}()

	mod.log.Infov(2, "using socket %s", mod.socket.LocalAddr())

	return tasks.Group(
		mod.discover,
		mod.announce,
		NewAPIService(mod),
	).Run(ctx)
}

func (mod *Module) Broadcast(flags ...string) error {
	_, err := mod.announce.announceWithFlag(flags...)
	return err
}

func (mod *Module) List() []*presence.Presence {
	var list []*presence.Presence

	for _, ad := range mod.discover.RecentAds() {
		list = append(list, &presence.Presence{
			Identity: ad.Identity,
			Alias:    ad.Alias,
			Flags:    ad.Flags,
		})
	}

	return list
}

func (mod *Module) SetVisible(b bool) error {
	if mod.visible.CompareAndSwap(!b, b) {
		if b {
			mod.AddHookAdOut(NewFlagOnce(mod, presence.DiscoverFlag))
		}
		select {
		case mod.announce.v <- b:
		default:
		}
	}

	return nil
}

func (mod *Module) Visible() bool {
	return mod.visible.Load()
}

func (mod *Module) AddHookAdOut(filter presence.AdOutHook) error {
	return mod.outFilters.Add(filter)
}

func (mod *Module) RemoveHookAdOut(filterFunc presence.AdOutHook) error {
	return mod.outFilters.Remove(filterFunc)
}

func (mod *Module) myAlias() string {
	a, _ := mod.Dir.GetAlias(mod.node.Identity())
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
