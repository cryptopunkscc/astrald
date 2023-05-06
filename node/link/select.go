package link

type SelectFunc func(current *Link, next *Link) *Link

// Select selects a link from the link array using the provided select function.
func Select(links []*Link, selectFunc SelectFunc) (selected *Link) {
	for _, next := range links {
		selected = selectFunc(selected, next)
	}
	return
}

// LowestPing selects the link with the lowest ping
func LowestPing(current *Link, next *Link) *Link {
	if current == nil {
		return next
	}

	if next.Health().LastRTT() < current.Health().LastRTT() {
		return next
	}

	return current
}

// MostRecent selects the link with the shortest idle duration
func MostRecent(current *Link, next *Link) *Link {
	if current == nil {
		return next
	}

	if next.activity.Idle() < current.activity.Idle() {
		return next
	}

	return current
}

// HighestPriority selects the link with the highest priority
func HighestPriority(current *Link, next *Link) *Link {
	if current == nil {
		return next
	}

	if next.Priority() > current.Priority() {
		return next
	}

	return current
}
