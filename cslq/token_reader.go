package cslq

import (
	"encoding/hex"
	"errors"
	"fmt"
	"io"
)

type TokenReader interface {
	Read() (Token, error)
}

// StreamTokenizer represents a tokenizer that reads bytes from the provided io.Reader and parses them
// into tokens.
type StreamTokenizer struct {
	stack *ByteStackReader
}

// TokenizeStream returns a new instance of a StreamTokenizer over the provided io.Reader.
func TokenizeStream(r io.Reader) *StreamTokenizer {
	return &StreamTokenizer{stack: NewByteStackReader(r)}
}

// Read reads and returns the next Token from the io.Reader
func (r *StreamTokenizer) Read() (Token, error) {
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
		return TokenStructStart{}, nil

	case '}':
		return TokenStructEnd{}, nil

	case '[':
		return TokenArrayStart{}, nil

	case ']':
		return TokenArrayEnd{}, nil

	case '<':
		return TokenConstStart{}, nil

	case '>':
		return TokenConstEnd{}, nil

	case 'c':
		return TokenUint8{}, nil

	case 's':
		return TokenUint16{}, nil

	case 'l':
		return TokenUint32{}, nil

	case 'q':
		return TokenUint64{}, nil

	case 'v':
		return TokenInterface{}, nil

	case 'x':
		var (
			err   error
			bytes [2]byte
		)

		if bytes[0], err = r.stack.Pop(); err != nil {
			return nil, err
		}
		if bytes[1], err = r.stack.Pop(); err != nil {
			return nil, err
		}

		if dec, err := hex.DecodeString(string(bytes[:])); err != nil {
			return nil, errors.New("invalid byte literal")
		} else {
			return TokenByteLiteral(dec[0]), nil
		}

	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		n := uint64(b - '0')

		for {
			c, err := r.stack.Pop()
			if err != nil {
				return TokenNumberLiteral(n), nil
			}

			if (c < '0') || (c > '9') {
				r.stack.Push(c)
				return TokenNumberLiteral(n), nil
			}

			n = n*10 + uint64(c-'0')
		}
	default:
		return nil, fmt.Errorf("invalid format character: %c", b)
	}
}
