package admin

type Arrow bool
type DoubleArrow bool

func (b Arrow) String() string {
	if b {
		return "->"
	}
	return "<-"
}

func (b DoubleArrow) String() string {
	if b {
		return "=>"
	}
	return "<="
}
