package nodes

type TracedLink struct {
	*Link
}

func NewTracedLink(link *Link, onClose func(error)) *TracedLink {
	tl := &TracedLink{Link: link}
	if onClose != nil {
		go func() {
			<-link.Done()
			onClose(link.Err())
		}()
	}
	return tl
}
