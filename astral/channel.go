package astral

import (
	"bufio"
	"bytes"
	encoding2 "encoding"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
)

type Channel struct {
	*Blueprints
	format string
	rw     io.ReadWriter
	jenc   *json.Encoder
	jdec   *json.Decoder
	bufr   *bufio.Reader
}

type jsonDecodeAdapter struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type jsonEncodeAdapter struct {
	Type    string `json:"type"`
	Payload any    `json:"payload"`
}

func NewChannel(rw io.ReadWriter, format string) *Channel {
	ch := &Channel{
		rw:         rw,
		Blueprints: ExtractBlueprints(rw),
		format:     format,
	}

	switch format {
	case "json":
		ch.jenc = json.NewEncoder(rw)
		ch.jdec = json.NewDecoder(rw)

	case "text", "textp", "astral", "":
		ch.bufr = bufio.NewReader(rw)
	}

	return ch
}

func (ch *Channel) Read() (obj Object, err error) {
	switch ch.format {
	case "", "bin":
		var frame Bytes16

		_, err = frame.ReadFrom(ch.bufr)
		if err != nil {
			return
		}

		obj, _, err = ch.Blueprints.Read(bytes.NewReader(frame), false)
		return

	case "json":
		var jsonObj jsonDecodeAdapter

		err = ch.jdec.Decode(&jsonObj)
		if err != nil {
			return
		}

		obj = ch.Blueprints.Make(jsonObj.Type)
		if obj == nil {
			obj = &RawObject{}
		}

		err = json.Unmarshal(jsonObj.Payload, &obj)
		return

	case "text", "textp":
		line, err := ch.bufr.ReadString('\n')
		if err != nil {
			return nil, err
		}
		line, _ = strings.CutSuffix(line, "\n")

		typ, text, err := splitTypeAndPayload(line)
		if err != nil {
			return nil, fmt.Errorf("invalid text format: %w", err)
		}

		obj = ch.Blueprints.Make(typ)
		u, ok := obj.(encoding2.TextUnmarshaler)
		if !ok {
			return nil, errors.New("object does not implement text decoding")
		}

		err = u.UnmarshalText([]byte(text))

		return obj, err
	}

	return nil, errors.New("unsupported channel format: " + ch.format)
}

func (ch *Channel) Write(obj Object) (err error) {
	switch ch.format {
	case "", "bin":
		var frame = &bytes.Buffer{}
		_, _ = String8(obj.ObjectType()).WriteTo(frame)

		_, err = obj.WriteTo(frame)
		if err != nil {
			return
		}

		_, err = Bytes16(frame.Bytes()).WriteTo(ch.rw)
		return

	case "json":
		err = ch.jenc.Encode(&jsonEncodeAdapter{
			Type:    obj.ObjectType(),
			Payload: obj,
		})
		return

	case "text", "textp":
		m, ok := obj.(encoding2.TextMarshaler)
		if !ok {
			return errors.New("object does not implement text encoding")
		}

		text, err := m.MarshalText()
		if err != nil {
			return err
		}

		switch ch.format {
		case "text":
			_, err = fmt.Fprintf(ch.rw, "#[%s] %s\n", obj.ObjectType(), string(text))
		case "textp":
			_, err = fmt.Fprintf(ch.rw, "%s\n", string(text))
		}

		return err
	}

	return errors.New("unsupported channel format: " + ch.format)
}

func (ch *Channel) Close() error {
	if c, ok := ch.rw.(io.Closer); ok {
		return c.Close()
	}
	return errors.New("transport doesn't support closing")
}

func (ch *Channel) Transport() io.ReadWriter {
	return ch.rw
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
