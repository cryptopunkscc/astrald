package cslq

import "errors"

type TokenStructStart struct{}
type TokenStructEnd struct{}

func (t TokenStructStart) Compile(tokens TokenReader) (Op, error) {
	var op = make(OpStruct, 0)

	// read types until we encounter a TokenStructEnd ("}")
	for {
		// get next token
		nextToken, err := tokens.Read()
		if err != nil {
			return nil, err
		}

		// we're done if it's StructEnd
		if _, ok := nextToken.(TokenStructEnd); ok {
			return op, nil
		}

		// compile it otherwise
		nextOp, err := nextToken.Compile(tokens)
		if err != nil {
			return nil, err
		}
		op = append(op, nextOp)
	}
}

func (s TokenStructEnd) Compile(_ TokenReader) (Op, error) {
	return nil, errors.New("unexpected StructEnd token")
}
