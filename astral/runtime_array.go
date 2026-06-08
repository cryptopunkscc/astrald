package astral

import (
	"bytes"
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
// registered. Element resolution uses defaultBlueprints; the codec path uses
// newRuntimeArrayWith internally for per-call registries.
func NewRuntimeArray(typeName string, length uint32) (*RuntimeArray, error) {
	return newRuntimeArrayWith(defaultBlueprints, typeName, length)
}

// NewRuntimeArrayWith is the bps-aware constructor for callers that hold a *Blueprints (e.g.
// a per-call registry built around DefaultBlueprints). Same shape as NewRuntimeArray; element
// resolution consults bps instead of the package default. Use when populating a RuntimeObject
// field whose ArraySpec.Type names a Blueprint registered only in bps.
func NewRuntimeArrayWith(bps *Blueprints, typeName string, length uint32) (*RuntimeArray, error) {
	return newRuntimeArrayWith(bps, typeName, length)
}

// newRuntimeArrayWith is the internal bps-aware constructor used by readField when decoding
// through a custom registry (WithBlueprints).
func newRuntimeArrayWith(bps *Blueprints, typeName string, length uint32) (*RuntimeArray, error) {
	et, err := resolveElemType(bps, typeName)
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
	bps := blueprintsFromReader(r)
	if !isRuntimeBlueprintType(bps, a.elemName) {
		return arrayValue{Value: a.ptr.Elem()}.ReadFrom(r)
	}
	arr := a.ptr.Elem()
	// why: parity with the JSON slow path and ptrValue.ReadFrom on nil-flag — without this
	// reset, a re-decode that sees nil-flag=0 silently keeps the previous element pointer
	// (readRuntimeBlueprintPtr doesn't touch the slot when the flag is 0).
	arr.SetZero()
	var n int64
	for i := 0; i < arr.Len(); i++ {
		m, err := readRuntimeBlueprintPtr(r, bps, a.elemName, arr.Index(i))
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

// UnmarshalJSON: see RuntimeSlice.UnmarshalJSON note — JSON path uses defaultBlueprints only.
// `null` zeroes the array on both fast and slow paths, mirroring map/slice JSON behaviour.
func (a *RuntimeArray) UnmarshalJSON(data []byte) error {
	if bytes.Equal(bytes.TrimSpace(data), jsonNull) {
		a.ptr.Elem().SetZero()
		return nil
	}
	if !isRuntimeBlueprintType(defaultBlueprints, a.elemName) {
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
