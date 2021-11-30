package link

type FilterFunc func(link *Link) bool

func Filter(links <-chan *Link, filterFunc FilterFunc) <-chan *Link {
	var out = make(chan *Link, len(links))
	defer close(out)

	for link := range links {
		if filterFunc(link) {
			out <- link
		}
	}

	return out
}

func OnlyNetwork(network string) FilterFunc {
	return func(link *Link) bool {
		return link.Network() == network
	}
}

// Networks reads all links from the channel and returns a list of networks without duplicates
func Networks(links <-chan *Link) []string {
	list := make([]string, 0)
	dup := make(map[string]struct{})

	for l := range links {
		if _, found := dup[l.Network()]; !found {
			dup[l.Network()] = struct{}{}
			list = append(list, l.Network())
		}
	}

	return list
}
