package service

import (
	"github.com/cryptopunkscc/astrald/app/warpdrive/api"
	"github.com/cryptopunkscc/astrald/app/warpdrive/platform/android"
	"github.com/cryptopunkscc/astrald/app/warpdrive/platform/desktop"
	"github.com/cryptopunkscc/astrald/app/warpdrive/platform/stub"
	"github.com/cryptopunkscc/astrald/app/warpdrive/storage/memory"
)

type Notify struct {
	Core          *api.Core
	notify        api.Notify
	notifications chan api.Notification
}

func (srv *Notify) Init() {
	srv.notifications = make(chan api.Notification, 128)
	srv.Core.Notify = srv.notifications
	srv.notify = srv.newNotify()
}

func (srv *Notify) newNotify() api.Notify {
	switch srv.Core.Platform {
	case api.PlatformDesktop:
		return desktop.Notifier{}
	case api.PlatformAndroid:
		return android.Notifier{}.Init()
	default:
		return stub.Notifier{}
	}
}

func (srv *Notify) Start() {
	for n := range srv.notifications {
		srv.dispatch(n)
	}
}

func (srv *Notify) dispatch(n api.Notification) {
	switch n.Status.Status {
	case api.StatusAdded:
		p := memory.Peers(*srv.Core).Get()[n.Peer]
		if n.Incoming && p.Mod == api.PeerModAsk {
			srv.notify.New(n)
		}
	case api.StatusProgress:
		srv.notify.Progress(n)
	case
		api.StatusFailed,
		api.StatusRejected,
		api.StatusCompleted:
		srv.notify.Finish(n)
	}
}
