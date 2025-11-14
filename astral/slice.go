package astral

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"reflect"
)

var _ Object = &Slice[Object]{}

type Slice[T Object] struct {
	Elem  *[]T
	Typed bool
}

type ObjectTyper interface {
	ObjectType() string
}

func WrapSlice[T Object](elem *[]T) *Slice[T] {
	return &Slice[T]{
		Elem:  elem,
		Typed: reflect.TypeOf((*T)(nil)).Elem().Kind() == reflect.Interface,
	}
}

func (Slice[T]) ObjectType() string {
	return ""
}

func (a Slice[T]) WriteTo(w io.Writer) (n int64, err error) {
	v := *a.Elem

	// write length
	n, err = Uint32(len(v)).WriteTo(w)
	if err != nil {
		return
	}

	var m int64
	for _, v := range v {
		var buf = &bytes.Buffer{}

		if a.Typed {
			_, err = Write(buf, v)
		} else {
			_, err = v.WriteTo(buf)
		}
		if err != nil {
			return
		}

		m, err = Uint32(buf.Len()).WriteTo(w)
		n += m
		if err != nil {
			return
		}

		var j int
		j, err = w.Write(buf.Bytes())
		n += int64(j)
		if err != nil {
			return
		}
	}

	return
}

func (a *Slice[T]) ReadFrom(r io.Reader) (n int64, err error) {
	var l Uint32
	n, err = l.ReadFrom(r)
	if err != nil {
		return
	}

	v := make([]T, l)

	var m int64
	for i := 0; i < int(l); i++ {
		var e T
		typ := reflect.TypeOf((*T)(nil)).Elem()
		if typ.Kind() == reflect.Pointer {
			e = reflect.New(reflect.TypeOf(e).Elem()).Interface().(T)
		}

		// read element length
		var elementLen Uint32
		m, err = elementLen.ReadFrom(r)
		n += m
		if err != nil {
			return
		}

		// read element bytes
		var buf = make([]byte, elementLen)
		var j int
		j, err = io.ReadFull(r, buf)
		n += int64(j)
		if err != nil {
			return
		}

		if a.Typed {
			var o Object
			var ok bool

			o, _, err = ExtractBlueprints(r).Read(bytes.NewReader(buf))
			e, ok = o.(T)
			if !ok {
				err = errors.New("typecast failed")
			}
		} else {
			_, err = e.ReadFrom(bytes.NewReader(buf))
		}
		if err != nil {
			return
		}
		v[i] = e
	}
	*a.Elem = v

	return
}

func (a *Slice[T]) MarshalJSON() ([]byte, error) {
	v := *a.Elem

	var list []JSONEncodeAdapter
	for _, o := range v {
		list = append(list, JSONEncodeAdapter{
			Type:   o.ObjectType(),
			Object: o,
		})
	}

	return json.Marshal(list)
}

func (a *Slice[T]) UnmarshalJSON(bytes []byte) error {
	if a.Elem == nil {
		a.Elem = new([]T)
	}

	var jlist []JSONDecodeAdapter
	if err := json.Unmarshal(bytes, &jlist); err != nil {
		return err
	}

	result := make([]T, len(jlist))
	for i, j := range jlist {
		obj := DefaultBlueprints.Make(j.Type)
		if obj == nil {
			// Not recognized object -> RawObject
			obj = &RawObject{}
		}

		var err error
		switch {
		case j.Object != nil:
			err = json.Unmarshal(j.Object, obj)
			if err != nil {
				return err
			}
		case j.Payload != nil:
			raw := &RawObject{
				Type:    j.Type,
				Payload: j.Payload,
			}
			obj, err = DefaultBlueprints.Refine(raw)
			if err != nil {
				obj = raw
			}
		}

		casted, ok := obj.(T)
		if !ok {
			return errors.New("Slice.UnmarshalJSON: typecast failed")
		}

		result[i] = casted
	}

	*a.Elem = result
	return nil
}
