package astral

import (
	"fmt"
	"io"
	"reflect"
)

var _ Object = (*runtimeSlice)(nil)

// runtimeSlice is the carrier for SliceSpec fields in runtime Blueprints. It owns a typed
// native slice (heterogeneous []Object or homogeneous []*T) and delegates encoding to the
// reflective sliceValue codec for byte-identical parity with Objectify on a native slice
// field.
type runtimeSlice struct {
	ptr reflect.Value // *[]elemType, always addressable
}

// astral:blueprint-ignore
func (*runtimeSlice) ObjectType() string { return "slice" }

// newRuntimeSlice returns a runtimeSlice whose element type is determined from a SliceSpec
// type name. An empty typeName means heterogeneous (element type = Object interface).
// Returns ErrBlueprintNotFound if typeName is non-empty and not registered.
func newRuntimeSlice(typeName string) (*runtimeSlice, error) {
	et, err := resolveElemType(typeName)
	if err != nil {
		return nil, err
	}
	return &runtimeSlice{ptr: reflect.New(reflect.SliceOf(et))}, nil
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

func (s *runtimeSlice) WriteTo(w io.Writer) (int64, error) {
	return sliceValue{Value: s.ptr.Elem()}.WriteTo(w)
}

func (s *runtimeSlice) ReadFrom(r io.Reader) (int64, error) {
	return sliceValue{Value: s.ptr.Elem()}.ReadFrom(r)
}

func (s *runtimeSlice) Len() int {
	return s.ptr.Elem().Len()
}

func (s *runtimeSlice) At(i int) Object {
	return s.ptr.Elem().Index(i).Interface().(Object)
}

func (s *runtimeSlice) Append(o Object) error {
	elemT := s.ptr.Elem().Type().Elem()
	rv := reflect.ValueOf(o)
	if !rv.Type().AssignableTo(elemT) {
		return fmt.Errorf("runtime_slice: want %s, got %s", elemT, rv.Type())
	}
	s.ptr.Elem().Set(reflect.Append(s.ptr.Elem(), rv))
	return nil
}

// Each iterates in order; stop by returning a non-nil error.
func (s *runtimeSlice) Each(fn func(int, Object) error) error {
	sl := s.ptr.Elem()
	for i := 0; i < sl.Len(); i++ {
		if err := fn(i, sl.Index(i).Interface().(Object)); err != nil {
			return err
		}
	}
	return nil
}
