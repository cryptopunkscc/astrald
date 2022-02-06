package service

import (
	"github.com/cryptopunkscc/astrald/app/warpdrive/api"
	"github.com/cryptopunkscc/astrald/app/warpdrive/platform/android"
	"github.com/cryptopunkscc/astrald/app/warpdrive/platform/desktop"
	"github.com/cryptopunkscc/astrald/app/warpdrive/platform/stub"
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

func (srv *Notify) newNotify() (notify api.Notify) {
	switch srv.Core.Platform {
	case api.PlatformDesktop:
		notify = &desktop.Notifier{}
	case api.PlatformAndroid:
		notifier := &android.Notifier{}
		notifier = notifier.Init()
		if notifier != nil {
			notify = notifier
		}
	default:
		notify = &stub.Notifier{}
	}
	return
}

func (srv *Notify) Start() {
	for n := range srv.notifications {
		canNotify := srv.notify != nil
		if !canNotify {
			continue
		}
		srv.dispatch(n)
	}
}

func (srv *Notify) dispatch(n api.Notification) {
	switch n.Status.Status {
	case api.StatusAwaiting:
		if n.Incoming && n.Peer.Mod == api.PeerModAsk {
			srv.notify.New(n)
		}
	case api.StatusUpdated:
		srv.notify.Progress(n)
	case
		api.StatusFailed,
		api.StatusRejected,
		api.StatusCompleted:
		srv.notify.Finish(n)
	}
}
