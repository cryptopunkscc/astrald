package astral

import (
	"encoding/json"
	"fmt"
	"io"
	"reflect"
)

var _ Object = (*RuntimeArray)(nil)

// RuntimeArray is the carrier for ArraySpec fields in runtime Blueprints. It owns a typed
// native fixed-length array (heterogeneous [N]Object or homogeneous [N]*T) and delegates
// encoding to the reflective arrayValue codec for byte-identical parity with Objectify on a
// native array field. Length is part of the schema, so the wire format omits a count prefix.
//
// When elemName is a runtime Blueprint, ReadFrom takes a slow path that constructs each
// element via New(elemName); see RuntimeSlice for the same rationale.
type RuntimeArray struct {
	ptr      reflect.Value // *[Length]elemType, always addressable
	elemName string        // ArraySpec.Type — empty means heterogeneous
}

// astral:blueprint-ignore
func (*RuntimeArray) ObjectType() string { return "array" }

// NewRuntimeArray returns a RuntimeArray whose element type is determined from an ArraySpec
// type name and whose length matches the spec. An empty typeName means heterogeneous (element
// type = Object interface). Returns ErrBlueprintNotFound if typeName is non-empty and not
// registered.
func NewRuntimeArray(typeName string, length uint32) (*RuntimeArray, error) {
	et, err := resolveElemType(typeName)
	if err != nil {
		return nil, err
	}
	return &RuntimeArray{
		ptr:      reflect.New(reflect.ArrayOf(int(length), et)),
		elemName: typeName,
	}, nil
}

func (a *RuntimeArray) WriteTo(w io.Writer) (int64, error) {
	return arrayValue{Value: a.ptr.Elem()}.WriteTo(w)
}

func (a *RuntimeArray) ReadFrom(r io.Reader) (int64, error) {
	if !isRuntimeBlueprintType(a.elemName) {
		return arrayValue{Value: a.ptr.Elem()}.ReadFrom(r)
	}
	arr := a.ptr.Elem()
	var n int64
	for i := 0; i < arr.Len(); i++ {
		m, err := readRuntimeBlueprintPtr(r, a.elemName, arr.Index(i))
		n += m
		if err != nil {
			return n, err
		}
	}
	return n, nil
}

func (a *RuntimeArray) MarshalJSON() ([]byte, error) {
	return arrayValue{Value: a.ptr.Elem()}.MarshalJSON()
}

func (a *RuntimeArray) UnmarshalJSON(data []byte) error {
	if !isRuntimeBlueprintType(a.elemName) {
		return arrayValue{Value: a.ptr.Elem()}.UnmarshalJSON(data)
	}
	var raw []json.RawMessage
	err := json.Unmarshal(data, &raw)
	if err != nil {
		return err
	}
	arr := a.ptr.Elem()
	if len(raw) != arr.Len() {
		return fmt.Errorf("runtime_array: want %d elements, got %d", arr.Len(), len(raw))
	}
	arr.SetZero()
	for i, r := range raw {
		err := unmarshalRuntimeBlueprintPtr(r, a.elemName, arr.Index(i))
		if err != nil {
			return err
		}
	}
	return nil
}

func (a *RuntimeArray) Len() int {
	return a.ptr.Elem().Len()
}

func (a *RuntimeArray) At(i int) Object {
	return a.ptr.Elem().Index(i).Interface().(Object)
}

// Set assigns value to index i. The element's runtime type must be assignable to the carrier's
// element type. Out-of-range indices and shape mismatches are rejected.
func (a *RuntimeArray) Set(i int, o Object) error {
	if i < 0 || i >= a.ptr.Elem().Len() {
		return fmt.Errorf("runtime_array: index %d out of range [0,%d)", i, a.ptr.Elem().Len())
	}
	elemT := a.ptr.Elem().Type().Elem()
	rv := reflect.ValueOf(o)
	if !rv.IsValid() || !rv.Type().AssignableTo(elemT) {
		return fmt.Errorf("runtime_array: want %s, got %T", elemT, o)
	}
	a.ptr.Elem().Index(i).Set(rv)
	return nil
}

// Each iterates in order; stop by returning a non-nil error.
func (a *RuntimeArray) Each(fn func(int, Object) error) error {
	arr := a.ptr.Elem()
	for i := 0; i < arr.Len(); i++ {
		if err := fn(i, arr.Index(i).Interface().(Object)); err != nil {
			return err
		}
	}
	return nil
}
