package fmt

type ErrorView struct {
	err string
}

var _ View = ErrorView{}

func (e ErrorView) Render() string {
	return e.err
}

func (e ErrorView) String() string {
	return e.err
}
