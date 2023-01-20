package link

type SelectFunc func(current *Link, next *Link) *Link

func Select(ch <-chan *Link, selectFunc SelectFunc) (selected *Link) {
	for next := range ch {
		selected = selectFunc(selected, next)
	}
	return
}

func LowestRoundTrip(current *Link, next *Link) *Link {
	if current == nil {
		return next
	}

	if next.Ping() < current.Ping() {
		return next
	}

	return current
}

func MostRecent(current *Link, next *Link) *Link {
	if current == nil {
		return next
	}

	if next.Idle() < current.Idle() {
		return next
	}

	return current
}
