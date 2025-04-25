package astral

import (
	"bufio"
	"bytes"
	encoding2 "encoding"
	"encoding/json"
	"errors"
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
		
	case "text", "astral", "":
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

	case "text":
		line, err := ch.bufr.ReadString('\n')
		if err != nil {
			return nil, err
		}
		line, _ = strings.CutSuffix(line, "\n")

		if strings.HasPrefix(line, "^{") && strings.HasSuffix(line, "}") {
			line = line[2 : len(line)-1]
			typ, text, found := strings.Cut(line, ":")
			if !found {
				return nil, errors.New("invalid format")
			}

			obj = ch.Blueprints.Make(typ)
			u, ok := obj.(encoding2.TextUnmarshaler)
			if !ok {
				return nil, errors.New("object does not implement text decoding")
			}

			err = u.UnmarshalText([]byte(text))

			return obj, err
		}

		return (*String)(&line), nil
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

	case "text":
		m, ok := obj.(encoding2.TextMarshaler)
		if !ok {
			return errors.New("object does not implement text encoding")
		}

		text, err := m.MarshalText()
		if err != nil {
			return err
		}

		_, err = ch.rw.Write(append(text, '\n'))
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
