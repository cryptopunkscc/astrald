package cslq

import (
	"errors"
	"fmt"
	"reflect"
)

type TokenArrayStart struct{}
type TokenArrayEnd struct{}

func (token TokenArrayStart) Compile(tokens TokenReader) (Op, error) {
	var op OpArray

	// read array length
	lenTypeToken, err := tokens.Read()
	if err != nil {
		return nil, err
	}

	switch token := lenTypeToken.(type) {
	case TokenUint8, TokenUint16, TokenUint32, TokenUint64:
		op.LenOp, err = token.Compile(tokens)
		if err != nil {
			return nil, err
		}

	case TokenNumberLiteral:
		if token <= 0 {
			return nil, fmt.Errorf("fixed array length must be a positive integer")
		}
		op.FixedLen = int(token)

	default:
		return nil, fmt.Errorf("invalid array length type %s", reflect.TypeOf(lenTypeToken))
	}

	// consume the array end token ("]")
	if err := expectToken(tokens, TokenArrayEnd{}); err != nil {
		return nil, err
	}

	// fetch and compile element op
	elemToken, err := tokens.Read()
	if err != nil {
		return nil, err
	}

	op.ElemOp, err = elemToken.Compile(tokens)
	if err != nil {
		return nil, err
	}

	return op, nil
}

func (a TokenArrayEnd) Compile(_ TokenReader) (Op, error) {
	return nil, errors.New("unexpected ArrayEnd token")
}
