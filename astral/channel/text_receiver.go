package channel

import (
	"bufio"
	"encoding"
	"fmt"
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
	var line, objectType, text string

	// read the line
	line, err = r.br.ReadString('\n')
	if err != nil {
		return nil, err
	}
	line, _ = strings.CutSuffix(line, "\n")

	// parse type and text
	objectType, text, err = splitTypeAndPayload(line)
	if err != nil {
		return nil, err
	}

	obj = r.bp.Make(objectType)
	if obj == nil {
		return nil, astral.ErrBlueprintNotFound{Type: objectType}
	}

	u, ok := obj.(encoding.TextUnmarshaler)
	if !ok {
		return nil, ErrTextUnsupported
	}

	err = u.UnmarshalText([]byte(text))

	return
}

func splitTypeAndPayload(line string) (string, string, error) {
	endIdx := strings.Index(line, "]")
	if endIdx == -1 {
		return "", "", fmt.Errorf("invalid format: missing closing bracket")
	}

	if !strings.HasPrefix(line, "#[") {
		return "", "", fmt.Errorf("invalid format: must start with '#['")
	}

	typeName := line[2:endIdx]
	if typeName == "" {
		return "", "", fmt.Errorf("invalid format: type name is empty")
	}

	payload := strings.TrimSpace(line[endIdx+1:])

	return typeName, payload, nil
}
