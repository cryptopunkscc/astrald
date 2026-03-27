package objects

// ErrTagUnsupported - a tag in the search query is unsupported by the describer
type ErrTagUnsupported struct {
	tagName string
}

func NewErrTagUnsupported(tagName string) ErrTagUnsupported {
	return ErrTagUnsupported{tagName: tagName}
}

func (e ErrTagUnsupported) Error() string {
	return "tag " + e.tagName + " is unsupported"
}

func (e ErrTagUnsupported) Is(target error) bool {
	_, ok := target.(ErrTagUnsupported)
	return ok
}
