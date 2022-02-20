package service

import (
	"github.com/cryptopunkscc/astrald/app/warpdrive/api"
	"time"
)

type Notify struct {
	Core          api.Core
	Notify        api.Notify
	notifications chan api.Notification
}

func (srv *Notify) Init() {
	srv.notifications = make(chan api.Notification, 128)
	srv.Core.Notify = srv.notifications
}

func (srv *Notify) Start() {
	debounce := int64(500)
	lastUpdate := int64(0)
	for n := range srv.notifications {
		canNotify := srv.Notify != nil
		if !canNotify {
			continue
		}
		if n.Status.Status == api.StatusUpdated &&
			time.Now().UnixMilli() < lastUpdate+debounce {
			continue
		}
		srv.dispatch(n)
		lastUpdate = time.Now().UnixMilli()
	}
}

func (srv *Notify) dispatch(n api.Notification) {
	switch n.Status.Status {
	case api.StatusAwaiting:
		if n.Incoming && n.Peer.Mod == api.PeerModAsk {
			srv.Notify.New(n)
		}
	case api.StatusUpdated:
		srv.Notify.Progress(n)
	case
		api.StatusFailed,
		api.StatusRejected,
		api.StatusCompleted:
		srv.Notify.Finish(n)
	}
}
