package nodes

import "sync/atomic"

type TracedLink struct {
	*Link
	OnClose func()
	closed  atomic.Bool
}

func NewTracedLink(link *Link, onClose func()) *TracedLink {
	return &TracedLink{Link: link, OnClose: onClose}
}

func (t *TracedLink) Close() error {
	if !t.closed.CompareAndSwap(false, true) {
		return nil
	}
	if t.OnClose != nil {
		t.OnClose()
	}
	return t.Link.Close()
}
