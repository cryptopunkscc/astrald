package astral

var _ Object = &EOS{}

// EOS is the end-of-stream sentinel object that marks the end of an object stream.
type EOS struct {
	EmptyObject
}

func (E EOS) ObjectType() string {
	return "eos"
}

func init() {
	Add(&EOS{})
}
