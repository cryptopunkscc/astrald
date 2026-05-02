package nodes

type TracedLink struct {
	*Link
}

func NewTracedLink(link *Link, onClose func()) *TracedLink {
	tl := &TracedLink{Link: link}
	if onClose != nil {
		go func() {
			<-link.Done()
			onClose()
		}()
	}
	return tl
}
