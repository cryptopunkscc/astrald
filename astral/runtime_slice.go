package astral

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
)

var _ Object = (*RuntimeSlice)(nil)

// RuntimeSlice is the carrier for SliceSpec fields in runtime Blueprints. It owns a typed
// native slice (heterogeneous []Object or homogeneous []*T) and delegates encoding to the
// reflective sliceValue codec for byte-identical parity with Objectify on a native slice
// field.
//
// When elemName is a runtime Blueprint, the generic reflective codec would allocate elements
// via reflect.New — which produces unbound *RuntimeObject{bp:nil} that silently decode to 0
// bytes — so ReadFrom takes a slow path that constructs each element via New(elemName).
type RuntimeSlice struct {
	ptr      reflect.Value // *[]elemType, always addressable
	elemName string        // SliceSpec.Type — empty means heterogeneous
}

// astral:blueprint-ignore
func (*RuntimeSlice) ObjectType() string { return "slice" }

// NewRuntimeSlice returns a RuntimeSlice whose element type is determined from a SliceSpec
// type name. An empty typeName means heterogeneous (element type = Object interface).
// Returns ErrBlueprintNotFound if typeName is non-empty and not registered.
func NewRuntimeSlice(typeName string) (*RuntimeSlice, error) {
	et, err := resolveElemType(typeName)
	if err != nil {
		return nil, err
	}
	return &RuntimeSlice{
		ptr:      reflect.New(reflect.SliceOf(et)),
		elemName: typeName,
	}, nil
}

// resolveElemType maps a spec type name to its reflect.Type. Empty name → Object interface.
// Concrete name → reflect.TypeOf(New(name)), which is the pointer prototype (e.g. *Uint32).
func resolveElemType(typeName string) (reflect.Type, error) {
	if typeName == "" {
		return reflect.TypeOf((*Object)(nil)).Elem(), nil
	}
	proto := New(typeName)
	if proto == nil {
		return nil, fmt.Errorf("%w: %s", ErrBlueprintNotFound, typeName)
	}
	return reflect.TypeOf(proto), nil
}

func (s *RuntimeSlice) WriteTo(w io.Writer) (int64, error) {
	return sliceValue{Value: s.ptr.Elem()}.WriteTo(w)
}

// ReadFrom decodes the slice. For elemName resolving to a runtime Blueprint, the generic
// reflective codec would silently no-op per element (see RuntimeSlice doc); the slow path
// constructs each element via New(elemName) so it carries its schema binding.
func (s *RuntimeSlice) ReadFrom(r io.Reader) (int64, error) {
	if !isRuntimeBlueprintType(s.elemName) {
		return sliceValue{Value: s.ptr.Elem()}.ReadFrom(r)
	}
	return s.readRuntimeBlueprintElements(r)
}

func (s *RuntimeSlice) readRuntimeBlueprintElements(r io.Reader) (int64, error) {
	var l uint32
	err := binary.Read(r, ByteOrder, &l)
	if err != nil {
		return 0, err
	}
	var n int64 = 4

	sl := reflect.MakeSlice(s.ptr.Elem().Type(), int(l), int(l))
	for i := 0; i < int(l); i++ {
		m, err := readRuntimeBlueprintPtr(r, s.elemName, sl.Index(i))
		n += m
		if err != nil {
			return n, err
		}
	}
	s.ptr.Elem().Set(sl)
	return n, nil
}

func (s *RuntimeSlice) MarshalJSON() ([]byte, error) {
	return sliceValue{Value: s.ptr.Elem()}.MarshalJSON()
}

// UnmarshalJSON mirrors ReadFrom: when elemName names a runtime Blueprint, allocate elements
// via New(elemName) so the underlying *RuntimeObject is bound before json.Unmarshal runs.
func (s *RuntimeSlice) UnmarshalJSON(data []byte) error {
	if !isRuntimeBlueprintType(s.elemName) {
		return sliceValue{Value: s.ptr.Elem()}.UnmarshalJSON(data)
	}
	var arr []json.RawMessage
	err := json.Unmarshal(data, &arr)
	if err != nil {
		return err
	}
	sl := reflect.MakeSlice(s.ptr.Elem().Type(), len(arr), len(arr))
	for i, raw := range arr {
		err := unmarshalRuntimeBlueprintPtr(raw, s.elemName, sl.Index(i))
		if err != nil {
			return err
		}
	}
	s.ptr.Elem().Set(sl)
	return nil
}

func (s *RuntimeSlice) Len() int {
	return s.ptr.Elem().Len()
}

func (s *RuntimeSlice) At(i int) Object {
	return s.ptr.Elem().Index(i).Interface().(Object)
}

func (s *RuntimeSlice) Append(o Object) error {
	elemT := s.ptr.Elem().Type().Elem()
	rv := reflect.ValueOf(o)
	if !rv.IsValid() || !rv.Type().AssignableTo(elemT) {
		return fmt.Errorf("runtime_slice: want %s, got %T", elemT, o)
	}
	s.ptr.Elem().Set(reflect.Append(s.ptr.Elem(), rv))
	return nil
}

// Each iterates in order; stop by returning a non-nil error.
func (s *RuntimeSlice) Each(fn func(int, Object) error) error {
	sl := s.ptr.Elem()
	for i := 0; i < sl.Len(); i++ {
		if err := fn(i, sl.Index(i).Interface().(Object)); err != nil {
			return err
		}
	}
	return nil
}
