package android

import (
	"fmt"
	"github.com/cryptopunkscc/astrald/app/warpdrive/api"
	"github.com/cryptopunkscc/astrald/mobile/android/service/notification/api"
	"log"
)

var _ api.Notify = &Notifier{}

type Notifier struct {
	notify.Api
	inChannel     notify.Channel
	outChannel    notify.Channel
	dispatch      chan<- []notify.Notification
	lastId        int
	notifications map[api.OfferId]notify.Notification
}

func (m Notifier) Init() *Notifier {
	if m.Api == nil {
		m.Api = notify.Client{}
	}
	m.inChannel = notify.Channel{
		Id:         "warpdrive-in",
		Name:       "Warp Drive incoming",
		Importance: notify.ImportanceMax,
	}
	m.outChannel = notify.Channel{
		Id:         "warpdrive-out",
		Name:       "Warp Drive outgoing",
		Importance: notify.ImportanceDefault,
	}
	m.notifications = map[api.OfferId]notify.Notification{}
	err := m.Create(m.inChannel)
	if err != nil {
		panic(err)
	}
	err = m.Create(m.outChannel)
	if err != nil {
		panic(err)
	}
	return &m
}

func (m *Notifier) New(an api.Notification) {
	title := "Sent"
	channel := m.outChannel
	size := int64(0)
	for _, info := range an.Files {
		size += info.Size
	}
	n := notify.Notification{
		Id:            m.nextId(),
		ChannelId:     channel.Id,
		Ongoing:       false,
		OnlyAlertOnce: true,
		AutoCancel:    true,
		Priority:      notify.PriorityMax,
		ContentText:   an.Files[0].Uri,
		SubText:       ByteCountSI(size),
		Number:        len(an.Files),
		ContentIntent: &notify.Intent{
			Uri: "warpdrive://" + string(an.Id),
		},
	}
	m.notifications[an.Id] = n
	if an.Incoming {
		channel = m.inChannel
		n.ChannelId = channel.Id
		title = "Received"
		n.ContentTitle = title
		err := m.Notify(n)
		if err != nil {
			log.Println("Cannot display notification", err)
		}
	}
}

func (m *Notifier) Progress(an api.Notification) {
	// TODO implement me
}

func (m *Notifier) Finish(an api.Notification) {
	//TODO implement me
}

func (m *Notifier) nextId() int {
	m.lastId++
	return m.lastId
}

func ByteCountSI(b int64) string {
	const unit = 1000
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB",
		float64(b)/float64(div), "kMGTPE"[exp])
}
