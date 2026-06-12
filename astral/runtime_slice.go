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
// Element resolution uses defaultBlueprints; for a per-call registry, the codec path uses
// newRuntimeSliceWith internally.
func NewRuntimeSlice(typeName string) (*RuntimeSlice, error) {
	return newRuntimeSliceWith(defaultBlueprints, typeName)
}

// NewRuntimeSliceWith is the bps-aware constructor for callers that hold a *Blueprints (e.g.
// a per-call registry built around DefaultBlueprints). Same shape as NewRuntimeSlice; element
// resolution consults bps instead of the package default. Use when populating a RuntimeObject
// field whose SliceSpec.Type names a Blueprint registered only in bps.
func NewRuntimeSliceWith(bps *Blueprints, typeName string) (*RuntimeSlice, error) {
	return newRuntimeSliceWith(bps, typeName)
}

// newRuntimeSliceWith is the internal bps-aware constructor used by readField when decoding
// through a custom registry (WithBlueprints). Same shape as NewRuntimeSlice, but resolveElemType
// consults the provided registry instead of the package default.
func newRuntimeSliceWith(bps *Blueprints, typeName string) (*RuntimeSlice, error) {
	et, err := resolveElemType(bps, typeName)
	if err != nil {
		return nil, err
	}
	return &RuntimeSlice{
		ptr:      reflect.New(reflect.SliceOf(et)),
		elemName: typeName,
	}, nil
}

// resolveElemType maps a spec type name to its reflect.Type. Empty name → Object interface.
// Concrete name → reflect.TypeOf(bps.New(name)), which is the pointer prototype (e.g. *Uint32).
func resolveElemType(bps *Blueprints, typeName string) (reflect.Type, error) {
	if typeName == "" {
		return reflect.TypeOf((*Object)(nil)).Elem(), nil
	}
	proto := bps.New(typeName)
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
// constructs each element via bps.New(elemName) so it carries its schema binding. bps is
// recovered from r via blueprintsFromReader so WithBlueprints flows through this frame.
func (s *RuntimeSlice) ReadFrom(r io.Reader) (int64, error) {
	bps := blueprintsFromReader(r)
	if !isRuntimeBlueprintType(bps, s.elemName) {
		return sliceValue{Value: s.ptr.Elem()}.ReadFrom(r)
	}
	return s.readRuntimeBlueprintElements(r, bps)
}

func (s *RuntimeSlice) readRuntimeBlueprintElements(r io.Reader, bps *Blueprints) (int64, error) {
	var l uint32
	err := binary.Read(r, ByteOrder, &l)
	if err != nil {
		return 0, err
	}
	var n int64 = 4

	sl := reflect.MakeSlice(s.ptr.Elem().Type(), int(l), int(l))
	for i := 0; i < int(l); i++ {
		m, err := readRuntimeBlueprintPtr(r, bps, s.elemName, sl.Index(i))
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
// note: JSON path uses defaultBlueprints — no plumbing exists to thread WithBlueprints
// through json.Unmarshal callbacks; tracked as a separate limitation.
func (s *RuntimeSlice) UnmarshalJSON(data []byte) error {
	if !isRuntimeBlueprintType(defaultBlueprints, s.elemName) {
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

// Append rejects o if its runtime type is not assignable to the carrier's element type;
// a homogeneous slice will not silently accept a foreign element.
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
