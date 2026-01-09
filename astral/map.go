package astral

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"github.com/cryptopunkscc/astrald/sig"
)

// Map is a structure that maps String16 keys to Object values. Keys can hold different types of objects.
type Map struct {
	sig.Map[string, Object]
}

// astral

func (Map) ObjectType() string { return "map" }

func (m Map) WriteTo(w io.Writer) (n int64, err error) {
	var i int64
	clone := m.Map.Clone()

	n, err = Uint32(len(clone)).WriteTo(w)
	if err != nil {
		return
	}

	for k, v := range clone {
		var data = &bytes.Buffer{}
		_, err = Encode(data, v)

		if err != nil {
			err = fmt.Errorf("pack object at %s: %w", k, err)
			return
		}

		i, err = String16(k).WriteTo(w)
		n += i
		if err != nil {
			return
		}

		i, err = Bytes32(data.Bytes()).WriteTo(w)
		n += i
		if err != nil {
			return
		}
	}

	return
}

func (m *Map) ReadFrom(r io.Reader) (n int64, err error) {
	var i int64
	var mapSize Uint32

	n, err = mapSize.ReadFrom(r)
	if err != nil {
		return
	}

	for _ = range mapSize {
		var key String16
		var data Bytes32
		var object Object

		i, err = key.ReadFrom(r)
		n += i
		if err != nil {
			return
		}

		i, err = data.ReadFrom(r)
		n += i
		if err != nil {
			return
		}

		object, _, err = Decode(bytes.NewReader(data))
		if err != nil {
			return
		}

		m.Map.Set(key.String(), object)
	}

	return
}

// json

func (m Map) MarshalJSON() ([]byte, error) {
	var err error
	var jmap = map[string]JSONAdapter{}

	_ = m.Each(func(k string, v Object) error {
		j := JSONAdapter{Type: v.ObjectType()}

		j.Object, err = json.Marshal(v)
		if err != nil {
			return err
		}

		jmap[k] = j

		return nil
	})

	return json.Marshal(jmap)
}

func (m *Map) UnmarshalJSON(bytes []byte) (err error) {
	var jmap map[string]JSONAdapter

	err = json.Unmarshal(bytes, &jmap)
	if err != nil {
		return
	}

	m.Map = sig.Map[string, Object]{}
	for k, jsonObj := range jmap {
		obj := New(jsonObj.Type)
		if obj == nil {
			return newErrBlueprintNotFound(jsonObj.Type)
		}

		if jsonObj.Object != nil {
			err = json.Unmarshal(jsonObj.Object, &obj)
			if err != nil {
				return
			}
		}

		m.Map.Set(k, obj)
	}

	return
}

func init() {
	Add(&Map{})
}
