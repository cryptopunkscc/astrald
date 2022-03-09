package cslq

import (
	"errors"
)

type Token interface {
	Compile(tokens TokenReader) (Op, error)
}

type TokenNumberLiteral uint64

func (TokenNumberLiteral) Compile(_ TokenReader) (Op, error) {
	return nil, errors.New("unexpected number literal")
}

type TokenInterface struct{}

func (i TokenInterface) Compile(_ TokenReader) (Op, error) {
	return OpInterface{}, nil
}
