package cslq

import "errors"

type TokenNumberLiteral uint64

func (TokenNumberLiteral) Compile(_ TokenReader) (Op, error) {
	return nil, errors.New("unexpected number literal")
}

type TokenByteLiteral uint8

func (t TokenByteLiteral) Compile(_ TokenReader) (Op, error) {
	return OpByte(t), nil
}
