package astral

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
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
	}

	return errors.New("unsupported channel format: " + ch.Format)
}

func (ch *Channel) Close() error {
	if c, ok := ch.rw.(io.Closer); ok {
		return c.Close()
	}
	return errors.New("transport doesn't support closing")
}
