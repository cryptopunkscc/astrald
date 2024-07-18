package astral

type SelectFunc func(current Link, next Link) Link

// SelectLink selects a link from the link array using the provided select function.
func SelectLink(links []Link, selectFunc SelectFunc) (selected Link) {
	for _, next := range links {
		selected = selectFunc(selected, next)
	}
	return
}
