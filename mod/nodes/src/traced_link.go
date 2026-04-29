package nodes

type TracedLink struct {
	*Link
	OnClose func()
}

func NewTracedLink(link *Link, onClose func()) *TracedLink {
	return &TracedLink{Link: link, OnClose: onClose}
}

func (t *TracedLink) Close() error {
	if t.OnClose != nil {
		t.OnClose()
	}

	return t.Link.Close()
}
