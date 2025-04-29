package astral

import (
	"bufio"
	"bytes"
	encoding2 "encoding"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/cryptopunkscc/astrald/streams"
	"io"
	"strings"
)

type Channel struct {
	*Blueprints
	fmtIn  string
	fmtOut string
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

// NewChannel creates a new channel
func NewChannel(rw io.ReadWriter) *Channel {
	return NewChannelFmt(rw, "", "")
}

// NewChannelFmt creates a channel with custom input/output formats
func NewChannelFmt(rw io.ReadWriter, fmtIn, fmtOut string) *Channel {
	ch := &Channel{
		rw:         rw,
		Blueprints: ExtractBlueprints(rw),
		fmtIn:      fmtIn,
		fmtOut:     fmtOut,
	}

	switch ch.fmtIn {
	case "json":
		ch.jdec = json.NewDecoder(rw)
	case "text", "text+", "astral", "":
		ch.bufr = bufio.NewReader(rw)
	}

	switch ch.fmtOut {
	case "json":
		ch.jenc = json.NewEncoder(rw)
	}

	return ch
}

func (ch *Channel) Read() (obj Object, err error) {
	switch ch.fmtIn {
	case "", "astral":
		var frame Bytes16

		_, err = frame.ReadFrom(ch.bufr)
		if err != nil {
			return
		}

		obj, _, err = ch.Blueprints.Read(bytes.NewReader(frame), false)

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

	case "text", "text+":
		var line, objectType, text string

		// read the line
		line, err = ch.bufr.ReadString('\n')
		if err != nil {
			return nil, err
		}
		line, _ = strings.CutSuffix(line, "\n")

		// parse type and text
		objectType, text, err = splitTypeAndPayload(line)
		if err != nil {
			return nil, fmt.Errorf("invalid text format: %w", err)
		}

		obj = ch.Blueprints.Make(objectType)
		if obj == nil {
			return nil, fmt.Errorf("unknown object type: %s", objectType)
		}

		u, ok := obj.(encoding2.TextUnmarshaler)
		if !ok {
			return nil, errors.New("object does not implement text decoding")
		}

		err = u.UnmarshalText([]byte(text))

	default:
		err = errors.New("unsupported input format: " + ch.fmtIn)
	}

	return
}

func (ch *Channel) ReadPayload(objectType string) (obj Object, err error) {
	obj = ch.Blueprints.Make(objectType)
	if obj == nil {
		return nil, errors.New("unknown object type")
	}

	switch ch.fmtIn {
	case "astral", "":
		var frame Bytes16

		_, err = frame.ReadFrom(ch.bufr)
		if err != nil {
			return
		}

		_, err = obj.ReadFrom(bytes.NewReader(frame))

	case "json":
		err = ch.jdec.Decode(&obj)

	case "text", "text+":
		u, ok := obj.(encoding2.TextUnmarshaler)
		if !ok {
			return nil, errors.New("object does not implement text decoding")
		}

		var line string

		line, err = ch.bufr.ReadString('\n')
		if err != nil {
			return
		}

		err = u.UnmarshalText([]byte(line))

	default:
		err = errors.New("unsupported input format: " + ch.fmtIn)
	}

	return
}

func (ch *Channel) WritePayload(obj Object) (err error) {
	switch ch.fmtOut {
	case "astral", "":
		var frame = &bytes.Buffer{}

		_, err = obj.WriteTo(frame)
		if err != nil {
			return
		}

		_, err = Bytes16(frame.Bytes()).WriteTo(ch.rw)

	case "json":
		err = ch.jenc.Encode(obj)

	case "text", "text+":
		m, ok := obj.(encoding2.TextMarshaler)
		if !ok {
			return errors.New("object does not implement text encoding")
		}

		var text []byte

		text, err = m.MarshalText()
		if err != nil {
			return err
		}

		_, err = fmt.Fprintf(ch.rw, "%s\n", string(text))

	default:
		err = errors.New("unsupported output format: " + ch.fmtOut)
	}

	return
}

func (ch *Channel) Write(obj Object) (err error) {
	switch ch.fmtOut {
	case "astral", "":
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

	case "text", "text+":
		m, ok := obj.(encoding2.TextMarshaler)
		if !ok {
			return errors.New("object does not implement text encoding")
		}

		text, err := m.MarshalText()
		if err != nil {
			return err
		}

		switch ch.fmtOut {
		case "text+":
			_, err = fmt.Fprintf(ch.rw, "#[%s] %s\n", obj.ObjectType(), string(text))
		case "text":
			_, err = fmt.Fprintf(ch.rw, "%s\n", string(text))
		}

		return err
	}

	return errors.New("unsupported channel format: " + ch.fmtOut)
}

func (ch *Channel) Close() error {
	if c, ok := ch.rw.(io.Closer); ok {
		return c.Close()
	}
	return errors.New("transport doesn't support closing")
}

func (ch *Channel) Transport() io.ReadWriter {
	return &streams.ReadWriteCloseSplit{
		Reader: ch.bufr,
		Writer: ch.rw,
		Closer: ch,
	}
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
