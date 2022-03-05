package cslq

import (
	"fmt"
	"io"
)

// TokenReader represents a tokenizer that reads bytes from the provided io.Reader and parses them
// into tokens.
type TokenReader struct {
	stack *ByteStackReader
}

// NewTokenReader returns a new instance of a TokenReader over the provided io.Reader.
func NewTokenReader(r io.Reader) *TokenReader {
	return &TokenReader{stack: NewByteStackReader(r)}
}

// Read reads and returns the next Token from the io.Reader
func (r *TokenReader) Read() (Token, error) {
	var b byte
	var err error

	// read next non-whitespace byte
	for {
		b, err = r.stack.Pop()
		if err != nil {
			return nil, err
		}
		switch b {
		case ' ', '\n', '\t':
			continue
		}
		break
	}

	switch b {
	case '{':
		return StructStartToken{}, nil

	case '}':
		return StructEndToken{}, nil

	case '[':
		return ArrayStartToken{}, nil

	case ']':
		return ArrayEndToken{}, nil

	case '<':
		return ExpectStartToken{}, nil

	case '>':
		return ExpectEndToken{}, nil

	case 'c':
		return Uint8Token{}, nil

	case 's':
		return Uint16Token{}, nil

	case 'l':
		return Uint32Token{}, nil

	case 'q':
		return Uint64Token{}, nil

	case 'v':
		return InterfaceToken{}, nil

	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		n := uint64(b - '0')

		for {
			c, err := r.stack.Pop()
			if err != nil {
				return NumberLiteralToken(n), nil
			}

			if (c < '0') || (c > '9') {
				r.stack.Push(c)
				return NumberLiteralToken(n), nil
			}

			n = n*10 + uint64(c-'0')
		}
	default:
		return nil, fmt.Errorf("invalid format character: %c", b)
	}
}
