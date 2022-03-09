package cslq

type Token interface {
	Compile(tokens TokenReader) (Op, error)
}

type TokenInterface struct{}

func (i TokenInterface) Compile(_ TokenReader) (Op, error) {
	return OpInterface{}, nil
}
