package link

type FilterFunc func(link *Link) bool

func Filter(links Stream, filterFunc FilterFunc) <-chan *Link {
	var out = make(chan *Link, len(links))
	defer close(out)

	for link := range links {
		if filterFunc(link) {
			out <- link
		}
	}

	return out
}

func Network(network string) FilterFunc {
	return func(link *Link) bool {
		return link.Network() == network
	}
}
