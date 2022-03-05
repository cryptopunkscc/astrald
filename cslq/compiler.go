package cslq

import (
	"errors"
	"fmt"
	"io"
	"reflect"
	"strings"
)

type Compiler struct {
	cache map[string]Format
}

var defaultCompiler = NewCompiler()

func NewCompiler() *Compiler {
	return &Compiler{cache: make(map[string]Format, 0)}
}

func Compile(s string) (Format, error) {
	return defaultCompiler.Compile(s)
}

func (c *Compiler) Compile(f string) (Format, error) {
	if cached, found := c.cache[f]; found {
		return cached, nil
	}

	format, err := c.compile(NewTokenReader(strings.NewReader(f)))
	if err != nil {
		return nil, err
	}
	c.cache[f] = format

	return format, err
}

func (c *Compiler) compile(tokens *TokenReader) (Format, error) {
	var s = make(Format, 0)

	for {
		nextOp, err := c.compileOp(tokens)
		if errors.Is(err, io.EOF) {
			return s, nil
		}
		if err != nil {
			return nil, err
		}
		s = append(s, nextOp)
	}
}

func (c *Compiler) compileOp(tokens *TokenReader) (Op, error) {
	token, err := tokens.Read()
	if err != nil {
		return nil, err
	}

	switch token.(type) {
	case ExpectStartToken:
		return c.compileExpect(tokens)
	case StructStartToken:
		return c.compileStruct(tokens)
	case ArrayStartToken:
		return c.compileArray(tokens)
	case Uint8Token:
		return OpUint8{}, nil
	case Uint16Token:
		return OpUint16{}, nil
	case Uint32Token:
		return OpUint32{}, nil
	case Uint64Token:
		return OpUint64{}, nil
	case InterfaceToken:
		return OpInterface{}, nil
	default:
		return nil, ErrUnexpectedToken{token}
	}
}

func (c *Compiler) compileArray(tokens *TokenReader) (OpArray, error) {
	var a OpArray

	// read array length
	lenTypeToken, err := tokens.Read()
	if err != nil {
		return OpArray{}, err
	}

	switch token := lenTypeToken.(type) {
	case Uint8Token:
		a.LenOp = OpUint8{}
	case Uint16Token:
		a.LenOp = OpUint16{}
	case Uint32Token:
		a.LenOp = OpUint32{}
	case Uint64Token:
		a.LenOp = OpUint64{}
	case NumberLiteralToken:
		if token <= 0 {
			return OpArray{}, fmt.Errorf("fixed array length must be a positive integer")
		}
		a.FixedLen = int(token)
	default:
		return OpArray{}, fmt.Errorf("invalid array length type %s", reflect.TypeOf(lenTypeToken))
	}

	// read the array end token ("]")
	endToken, err := tokens.Read()
	if err != nil {
		return OpArray{}, err
	}
	if _, ok := endToken.(ArrayEndToken); !ok {
		return OpArray{}, fmt.Errorf("expected ArrayEndToken, got %s", reflect.TypeOf(endToken))
	}

	// read element type
	a.ElemOp, err = c.compileOp(tokens)
	if err != nil {
		return OpArray{}, err
	}

	return a, nil
}

func (c *Compiler) compileStruct(tokens *TokenReader) (OpStruct, error) {
	var s = make(OpStruct, 0)

	// read types until we encounter a StructEndToken ("}")
	for {
		nextType, err := c.compileOp(tokens)
		if err, ok := err.(ErrUnexpectedToken); ok {
			if _, ok := err.Token.(StructEndToken); ok {
				return s, nil
			}
		}
		if err != nil {
			return nil, err
		}

		s = append(s, nextType)
	}
}

func (c *Compiler) compileExpect(tokens *TokenReader) (OpExpect, error) {
	var s = make(OpExpect, 0)

	// read types until we encounter a StructEndToken ("}")
	for {
		nextType, err := c.compileOp(tokens)
		if err, ok := err.(ErrUnexpectedToken); ok {
			if _, ok := err.Token.(ExpectEndToken); ok {
				return s, nil
			}
		}
		if err != nil {
			return nil, err
		}

		s = append(s, nextType)
	}
}
