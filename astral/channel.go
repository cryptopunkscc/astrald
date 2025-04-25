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
	Format string
	rw     io.ReadWriter
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
	return &Channel{
		rw:         rw,
		Blueprints: ExtractBlueprints(rw),
		Format:     format,
	}
}

func (ch *Channel) Read() (obj Object, err error) {
	switch ch.Format {
	case "", "bin":
		var frame Bytes16

		_, err = frame.ReadFrom(ch.rw)
		if err != nil {
			return
		}

		obj, _, err = ch.Blueprints.Read(bytes.NewReader(frame), false)
		return

	case "json":
		var jsonObj jsonDecodeAdapter

		err = json.NewDecoder(ch.rw).Decode(&jsonObj)
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
		var r = bufio.NewReader(ch.rw)
		line, err := r.ReadString('\n')
		if err != nil {
			return nil, err
		}

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

	return nil, errors.New("unsupported channel format: " + ch.Format)
}

func (ch *Channel) Write(obj Object) (err error) {
	switch ch.Format {
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
		err = json.NewEncoder(ch.rw).Encode(&jsonEncodeAdapter{
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

	return errors.New("unsupported channel format: " + ch.Format)
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
