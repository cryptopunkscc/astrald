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
	Elem     *[]T
	ElemType string
	Typed    bool
}

type ObjectTyper interface {
	ObjectType() string
}

func WrapSlice[T Object](elem *[]T) *Slice[T] {
	var elemType string
	// if slice is non-empty, use first element
	if elem != nil && len(*elem) > 0 {
		elemType = (*elem)[0].ObjectType()
	} else {
		// fallback: get type from T itself
		typ := reflect.TypeOf((*T)(nil)).Elem()
		elemType = typ.Name()
	}

	return &Slice[T]{
		Elem:     elem,
		Typed:    reflect.TypeOf((*T)(nil)).Elem().Kind() == reflect.Interface,
		ElemType: elemType,
	}
}

func (Slice[T]) ObjectType() string {
	return "slice"
}

func (a Slice[T]) WriteTo(w io.Writer) (n int64, err error) {
	v := *a.Elem

	// write length
	n, err = Uint32(len(v)).WriteTo(w)
	if err != nil {
		return
	}

	var m int64
	// write element type name once
	tn := []byte(a.ElemType)
	m, err = Uint32(len(tn)).WriteTo(w)
	n += m
	if err != nil {
		return
	}

	j, err := w.Write(tn)
	n += int64(j)
	if err != nil {
		return
	}

	for _, v := range v {
		var buf = &bytes.Buffer{}

		if a.Typed {
			_, err = Encode(buf, v)
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
	a.Elem = new([]T)

	var l Uint32
	n, err = l.ReadFrom(r)
	if err != nil {
		return
	}

	var m int64
	var tnLen Uint32
	m, err = tnLen.ReadFrom(r)
	n += m
	if err != nil {
		return
	}

	tn := make([]byte, tnLen)
	var j int
	j, err = io.ReadFull(r, tn)
	n += int64(j)
	if err != nil {
		return
	}

	a.ElemType = string(tn)
	if a.ElemType != "" {
		a.Typed = true
	}

	v := make([]T, l)

	for i := 0; i < int(l); i++ {
		var e T
		typ := reflect.TypeOf((*T)(nil)).Elem()
		if typ.Kind() == reflect.Pointer {
			e = reflect.New(reflect.TypeOf(e).Elem()).Interface().(T)
		}

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

		obj := New(a.ElemType)
		if obj == nil {
			return n, NewErrBlueprintNotFound(a.ElemType)
		}

		// read inner element from buffer
		_, err = obj.ReadFrom(bytes.NewReader(buf))
		if err != nil {
			return
		}

		// cast into T
		casted, ok := obj.(T)
		if !ok {
			return n, errors.New("Slice.ReadFrom: typecast failed")
		}

		e = casted

		v[i] = e
	}

	*a.Elem = v

	return
}

func (a *Slice[T]) MarshalJSON() ([]byte, error) {
	v := *a.Elem

	var list []JSONAdapter
	for _, o := range v {
		jsonBytes, err := json.Marshal(o)
		if err != nil {
			return nil, err
		}

		list = append(list, JSONAdapter{
			Type:   o.ObjectType(),
			Object: jsonBytes,
		})
	}

	return json.Marshal(list)
}

func (a *Slice[T]) UnmarshalJSON(bytes []byte) error {
	if a.Elem == nil {
		a.Elem = new([]T)
	}

	var jlist []JSONAdapter
	if err := json.Unmarshal(bytes, &jlist); err != nil {
		return err
	}

	result := make([]T, len(jlist))
	for i, j := range jlist {
		obj := New(j.Type)
		if obj == nil {
			return NewErrBlueprintNotFound(j.Type)
		}

		var err error
		if j.Object != nil {
			err = json.Unmarshal(j.Object, obj)
			if err != nil {
				return err
			}
		}

		// important: track ElemType
		if i == 0 {
			a.ElemType = j.Type
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

func init() {
	_ = Add(&Slice[Object]{})
}
