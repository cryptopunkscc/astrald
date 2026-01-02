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

	"github.com/cryptopunkscc/astrald/streams"
)

// Channel is deprecated. Use the channel package instead.
type Channel struct {
	*Blueprints
	fmtIn  string
	fmtOut string
	rw     io.ReadWriter
	jenc   *json.Encoder
	jdec   *json.Decoder
	bufr   *bufio.Reader
}

// Deprecated: use `channel.New()` instead.
func NewChannel(rw io.ReadWriter) *Channel {
	return NewChannelFmt(rw, "", "")
}

// Deprecated: use `channel.New()` instead.
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
	case "bin", "astral", "":
		// read the object type
		var objectType ObjectType
		_, err = objectType.ReadFrom(ch.bufr)
		if err != nil {
			return
		}

		if len(objectType) == 0 {
			obj = &Blob{}
		} else {
			obj = ch.Make(objectType.String())
			if obj == nil {
				return nil, errors.New("unknown object type: " + string(objectType))
			}
		}

		// read the object payload
		var buf Bytes32
		_, err = buf.ReadFrom(ch.bufr)
		if err != nil {
			return
		}

		// decode the payload
		_, err = obj.ReadFrom(bytes.NewReader(buf))

	case "json":
		var jsonObj JSONDecodeAdapter

		err = ch.jdec.Decode(&jsonObj)
		if err != nil {
			return
		}

		obj, err = ch.Blueprints.RefineJSON(&jsonObj)

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

func (ch *Channel) Write(obj Object) (err error) {
	switch ch.fmtOut {
	case "bin", "astral", "":
		// write the object type
		_, err = String8(obj.ObjectType()).WriteTo(ch.rw)
		if err != nil {
			return
		}

		// buffer the payload
		var buf = bytes.NewBuffer(nil)
		_, err = obj.WriteTo(buf)
		if err != nil {
			return
		}

		// write the buffer with 32-bit length prefix
		_, err = Bytes32(buf.Bytes()).WriteTo(ch.rw)

		return

	case "json":
		return WriteJSON(ch.rw, obj)

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
