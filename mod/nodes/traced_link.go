package nodes

type TracedLink struct {
	Link
	OnClose func()
}

func (t *TracedLink) Close() error {
	if t.OnClose != nil {
		t.OnClose()
	}
	return t.Link.Close()
}
