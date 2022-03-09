package cslq

import "errors"

type TokenConstStart struct{}
type TokenConstEnd struct{}

func (t TokenConstStart) Compile(tokens TokenReader) (Op, error) {
	var op = make(OpConst, 0)

	// read types until we encounter a ConstEnd token
	for {
		// get next token
		nextToken, err := tokens.Read()
		if err != nil {
			return nil, err
		}

		// we're done if it's StructEnd
		if _, ok := nextToken.(TokenConstEnd); ok {
			return op, nil
		}

		nextOp, err := nextToken.Compile(tokens)
		if err != nil {
			return nil, err
		}
		op = append(op, nextOp)
	}
}

func (t TokenConstEnd) Compile(_ TokenReader) (Op, error) {
	return nil, errors.New("unexpected ConstEnd token")
}
