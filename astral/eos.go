package astral

var _ Object = &EOS{}

type EOS struct {
	EmptyObject
}

func (E EOS) ObjectType() string {
	return "eos"
}

func init() {
	Add(&EOS{})
}
