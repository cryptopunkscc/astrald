package channel

import (
	"bufio"
	"bytes"
	"encoding"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/cryptopunkscc/astrald/astral"
)

type TextReceiver struct {
	r         io.Reader
	br        *bufio.Reader
	streamErr error
}

var _ Receiver = &TextReceiver{}

func NewTextReceiver(r io.Reader) *TextReceiver {
	return &TextReceiver{
		r:  r,
		br: bufio.NewReader(r),
	}
}

// AllowUnparsed is not honored on text streams. Although lines are self-delimited, the policy
// here is fail-fast: the first non-nil error latches and subsequent Receive() calls return it
// without touching the reader.
func (r *TextReceiver) Receive() (obj astral.Object, err error) {
	if r.streamErr != nil {
		return nil, r.streamErr
	}
	defer func() {
		if err != nil {
			r.streamErr = err
		}
	}()

	line, err := r.br.ReadString('\n')
	if err != nil {
		return nil, err
	}
	line, _ = strings.CutSuffix(line, "\n")

	parsed, err := ParseText(line)
	if err != nil {
		return nil, err
	}

	if parsed.Type == "" {
		obj = &astral.Blob{}
	} else {
		obj = astral.New(parsed.Type)
		if obj == nil {
			return nil, fmt.Errorf("%w: %w: %s", astral.ErrStreamCorrupted, astral.ErrBlueprintNotFound, parsed.Type)
		}
	}

	switch parsed.Enc {
	case "text":
		u, ok := obj.(encoding.TextUnmarshaler)
		if !ok {
			return nil, ErrTextUnsupported
		}

		err = u.UnmarshalText([]byte(parsed.Text))
		if err != nil {
			return nil, err
		}

	case "base64":
		var payload = make([]byte, base64.StdEncoding.DecodedLen(len(parsed.Text)))
		_, err = base64.StdEncoding.Decode(payload, []byte(parsed.Text))
		if err != nil {
			return nil, err
		}

		_, err = obj.ReadFrom(bytes.NewReader(payload))
		if err != nil {
			return nil, err
		}

	case "none":
		// no payload

	default:
		return nil, errors.New("unknown encoding")
	}

	return
}
