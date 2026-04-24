package fmt

type stringView string

func (v stringView) Render() string {
	return string(v)
}

func (v stringView) String() string {
	return string(v)
}
