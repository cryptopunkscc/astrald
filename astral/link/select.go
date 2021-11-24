package link

type SelectFunc func(current *Link, next *Link) *Link

func Select(stream Stream, selectFunc SelectFunc) (selected *Link) {
	for next := range stream {
		selected = selectFunc(selected, next)
	}
	return
}

func Fastest(current *Link, next *Link) *Link {
	if current == nil {
		return next
	}
	// prefer inet
	if next.Network() == "inet" {
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
