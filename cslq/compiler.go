package cslq

import (
	"errors"
	"io"
	"strings"
)

type Compiler struct {
	cache map[string]Format
}

var defaultCompiler = NewCompiler()

func NewCompiler() *Compiler {
	return &Compiler{cache: make(map[string]Format, 0)}
}

func Compile(pattern string) (Format, error) {
	return defaultCompiler.Compile(pattern)
}

func (c *Compiler) Compile(pattern string) (Format, error) {
	if cached, found := c.cache[pattern]; found {
		return cached, nil
	}

	format, err := c.compile(TokenizeStream(strings.NewReader(pattern)))
	if err != nil {
		return nil, err
	}

	c.cache[pattern] = format

	return format, err
}

func (c *Compiler) CompileTokens(tokens TokenReader) (Format, error) {
	return c.compile(tokens)
}

func (c *Compiler) compile(tokens TokenReader) (Format, error) {
	var format = make(Format, 0)

	for {
		// get next token
		nextToken, err := tokens.Read()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return format, nil
			}
			return nil, err
		}

		// compile the op
		nextOp, err := nextToken.Compile(tokens)
		if err != nil {
			return nil, err
		}

		format = append(format, nextOp)
	}
}
