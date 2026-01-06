package channel

import (
	"bufio"
	"bytes"
	"encoding"
	"encoding/base64"
	"errors"
	"io"
	"strings"

	"github.com/cryptopunkscc/astrald/astral"
)

type TextReceiver struct {
	r  io.Reader
	bp *astral.Blueprints
	br *bufio.Reader
}

var _ Receiver = &TextReceiver{}

func NewTextReceiver(r io.Reader) *TextReceiver {
	return &TextReceiver{
		r:  r,
		bp: astral.ExtractBlueprints(r),
		br: bufio.NewReader(r),
	}
}

func (r TextReceiver) Receive() (obj astral.Object, err error) {
	// read the line
	line, err := r.br.ReadString('\n')
	if err != nil {
		return nil, err
	}
	line, _ = strings.CutSuffix(line, "\n")

	// parse type and text
	parsed, err := ParseText(line)
	if err != nil {
		return nil, err
	}

	// make the object
	obj = r.bp.Make(parsed.Type)
	if obj == nil {
		return nil, astral.ErrBlueprintNotFound{Type: parsed.Type}
	}

	switch parsed.Enc {
	case " ":
		u, ok := obj.(encoding.TextUnmarshaler)
		if !ok {
			return nil, ErrTextUnsupported
		}

		err = u.UnmarshalText([]byte(parsed.Text))
		if err != nil {
			return nil, err
		}

	case ":":
		var payload = make([]byte, base64.StdEncoding.DecodedLen(len(parsed.Text)))
		_, err = base64.StdEncoding.Decode(payload, []byte(parsed.Text))
		if err != nil {
			return nil, err
		}

		_, err = obj.ReadFrom(bytes.NewReader(payload))
		if err != nil {
			return nil, err
		}

	default:
		return nil, errors.New("invalid encoding character")
	}

	return
}
