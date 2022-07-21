package android

import (
	"fmt"
	"github.com/cryptopunkscc/astrald/app/warpdrive/api"
	"github.com/cryptopunkscc/astrald/mobile/android/node/notify"
	"log"
	"strconv"
	"strings"
)

var _ api.Notify = (&Notifier{}).Notify

type Notifier struct {
	notify.Api
	notify        notify.Notify
	inChannel     notify.Channel
	outChannel    notify.Channel
	inGroup       notify.Notification
	outGroup      notify.Notification
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
	groupCommon := notify.Notification{
		SubText:       "Warp Drive",
		OnlyAlertOnce: true,
		AutoCancel:    true,
		Silent:        true,
		GroupSummary:  true,
	}
	m.inGroup = groupCommon
	m.inGroup.Id = m.nextId()
	m.inGroup.ChannelId = m.inChannel.Id
	m.inGroup.Group = "in"

	m.outGroup = groupCommon
	m.outGroup.Id = m.nextId()
	m.outGroup.ChannelId = m.outChannel.Id
	m.outGroup.Group = "out"

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
	m.notify = m.Api.Notifier()
	return m
}

func (m *Notifier) Notify(notifications []api.Notification) {
	for _, n := range notifications {
		switch n.Status {
		case api.StatusAwaiting:
			if n.In && n.Peer.Mod == api.PeerModAsk {
				m.create(n)
			}
		case api.StatusUpdated:
			m.progress(n)
		case
			api.StatusFailed,
			api.StatusRejected,
			api.StatusCompleted:
			m.finish(n)
		}
	}
	groups := []notify.Notification{m.inGroup, m.outGroup}
	var last *notify.Notification
	var arr []notify.Notification
	for _, group := range groups {
		for _, last = range m.notifications {
			if last.ChannelId == group.ChannelId {
				arr = append(arr, *last)
			}
		}
		if last != nil && last.ChannelId == group.ChannelId {
			arr = append(arr, group)
		}
	}
	m.notify <- arr
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
			Uri: "warpdrive://offer/" + string(an.Offer.Id),
		},
	}
	if an.In {
		n.Group = "in"
		n.Action = &notify.Action{
			Title: "download",
			Intent: &notify.Intent{
				Uri: "warpdrive://download/" + string(an.Offer.Id),
			},
		}
	} else {
		n.Group = "out"
	}
	if an.In {
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
	m.notifications[an.Offer.Id] = n
	return
}

func (m *Notifier) progress(an api.Notification) {
	n := m.notifications[an.Offer.Id]
	if n == nil {
		n = m.create(an)
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
	n.Action = nil
	m.notifications[an.Offer.Id] = n
}

func (m *Notifier) finish(an api.Notification) {
	n := m.notifications[an.Offer.Id]
	if n == nil {
		n = m.create(an)
	}
	n.Ongoing = false
	n.AutoCancel = true
	n.ContentTitle = titlePrefix(an) + " " + formatPeerName(an) + " " + an.OfferStatus.Status
	n.ContentText = fmt.Sprintf(
		"transferred %d/%d files with size %s",
		an.Index,
		len(an.Files),
		formatTransferredSize(an),
	)
	n.Progress = nil
	m.notifications[an.Offer.Id] = n
}

func titlePrefix(an api.Notification) string {
	if an.In {
		return "Downloading from"
	}
	return "Uploading to"
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
