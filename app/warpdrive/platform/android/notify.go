package android

import (
	"fmt"
	"github.com/cryptopunkscc/astrald/app/warpdrive/api"
	"github.com/cryptopunkscc/astrald/mobile/android/service/notification/api"
	"log"
	"strconv"
	"strings"
)

var _ api.Notify = &Notifier{}

type Notifier struct {
	notify.Api
	inChannel     notify.Channel
	outChannel    notify.Channel
	lastId        int
	notifications map[api.OfferId]*notify.Notification
}

func (m *Notifier) Init() *Notifier {
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
	m.notifications = map[api.OfferId]*notify.Notification{}
	err := m.Create(m.inChannel)
	if err != nil {
		log.Println("Cannot create incoming notification channel", err)
		return nil
	}
	err = m.Create(m.outChannel)
	if err != nil {
		log.Println("Cannot create outgoing notification channel", err)
		return nil
	}
	return m
}

func (m *Notifier) create(an api.Notification) (n *notify.Notification) {
	channel := m.outChannel
	n = &notify.Notification{
		Id:            m.nextId(),
		ChannelId:     channel.Id,
		Ongoing:       false,
		OnlyAlertOnce: true,
		AutoCancel:    true,
		Priority:      notify.PriorityMax,
		SubText:       "Warp Drive",
		Number:        len(an.Files),
		ContentIntent: &notify.Intent{
			Uri: "warpdrive://" + string(an.Offer.Id),
		},
	}
	m.notifications[an.Offer.Id] = n
	if an.Incoming {
		n.ChannelId = m.inChannel.Id
		peerName := an.Peer.Alias
		if peerName == "" {
			shortId := string(an.Peer.Id)[58:66]
			peerName = shortId[0:4] + "-" + shortId[4:8]
		}
		filename := ""
		size := int64(0)
		if len(an.Files) > 0 {
			filename = an.Files[0].Uri
		}
		for _, info := range an.Files {
			size += info.Size
			if !strings.HasPrefix(info.Uri, filename) {
				filename = ""
			}
		}
		title := peerName + " wants to share"
		text := ""
		if len(an.Files) > 1 {
			if filename != "" {
				title += " a directory " + filename
			}
			text = strconv.Itoa(len(an.Files)) + " files with summary size "
		} else {
			title += " a file " + filename
			text = "with size "
		}
		text += ByteCountSI(size)
		n.ContentTitle = title
		n.ContentText = text
	}
	return
}

func (m *Notifier) New(an api.Notification) {
	if n := m.create(an); an.In {
		err := m.Notify(*n)
		if err != nil {
			log.Println("Cannot display notification", err)
		}
	}
}

func formatPeerName(an api.Notification) (name string) {
	name = an.Peer.Alias
	if name == "" && len(an.Peer.Id) == 66 {
		shortId := string(an.Peer.Id)[58:66]
		name = shortId[0:4] + "-" + shortId[4:8]
	}
	if name == "" {
		name = "this device"
	}
	return
}

func titlePrefix(an api.Notification) string {
	if an.In {
		return "Downloading from"
	}
	return "Uploading to"
}

func (m *Notifier) Progress(an api.Notification) {
	n := m.notifications[an.Offer.Id]
	if n == nil {
		n = m.create(an)
		return
	}
	if an.Info == nil {
		log.Println("Cannot update progress for nil Info")
		return
	}
	n.Ongoing = true
	n.AutoCancel = false
	n.ContentTitle = titlePrefix(an) + " " + formatPeerName(an)
	n.ContentText = an.Uri + " " + ByteCountSI(an.Progress) + " / " + ByteCountSI(an.Size)
	n.Progress = &notify.Progress{
		Max:     int(an.Size),
		Current: int(an.Progress),
	}
	err := m.Notify(*n)
	if err != nil {
		log.Println("Cannot display notification", err)
	}
}

func (m *Notifier) Finish(an api.Notification) {
	n := m.notifications[an.Offer.Id]
	if n == nil {
		n = m.create(an)
		return
	}
	n.Ongoing = false
	n.AutoCancel = true
	n.ContentTitle = titlePrefix(an) + " " + formatPeerName(an) + " " + an.Status.Status
	n.ContentText = fmt.Sprintf(
		"transferred %d/%d files with size %s",
		an.Index,
		len(an.Files),
		formatTransferredSize(an),
	)
	n.Progress = nil
	err := m.Notify(*n)
	if err != nil {
		log.Println("Cannot display notification", err)
	}
}

func formatTransferredSize(an api.Notification) (str string) {
	str = ByteCountSI(sumSize(an))
	if an.Index < len(an.Files) {
		str += " / " + ByteCountSI(totalSize(an))
	}
	return
}

func totalSize(an api.Notification) (size int64) {
	for _, file := range an.Files {
		size += file.Size
	}
	return
}

func sumSize(an api.Notification) (size int64) {
	for i := 0; i < an.Index; i++ {
		size += an.Files[i].Size
	}
	size += an.Progress
	return
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
