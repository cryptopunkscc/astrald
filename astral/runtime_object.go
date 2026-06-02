package astral

import (
	"fmt"
	"io"
)

// RuntimeObject is the live encoder/decoder for a runtime-registered Blueprint. Field values are
// kept in declared order alongside their normalized astral types.
type RuntimeObject struct {
	bp     *Blueprint
	fields []fieldValue
}

type fieldValue struct {
	Name  string
	Value Object
}

// NewRuntimeObject returns a RuntimeObject whose fields are initialized to the zero value of each
// Spec.
func NewRuntimeObject(bp *Blueprint) *RuntimeObject {
	if bp == nil {
		return &RuntimeObject{}
	}
	ro := &RuntimeObject{bp: bp, fields: make([]fieldValue, 0, len(bp.Fields))}
	for _, f := range bp.Fields {
		ro.fields = append(ro.fields, fieldValue{Name: f.Name.String(), Value: specZero(f.Spec)})
	}
	return ro
}

// GetRuntimeObject returns a fresh RuntimeObject backed by this Blueprint.
func (bp *Blueprint) GetRuntimeObject() *RuntimeObject { return NewRuntimeObject(bp) }

// astral:blueprint-ignore
func (ro *RuntimeObject) ObjectType() string {
	if ro.bp == nil {
		return ""
	}
	return ro.bp.Type.String()
}

func (ro *RuntimeObject) WriteTo(w io.Writer) (n int64, err error) {
	if ro.bp == nil {
		return 0, nil
	}

	for i, f := range ro.bp.Fields {
		var m int64
		m, err = writeField(w, f.Spec, ro.fields[i].Value)
		n += m
		if err != nil {
			return
		}
	}
	return
}

func (ro *RuntimeObject) ReadFrom(r io.Reader) (n int64, err error) {
	if ro.bp == nil {
		return 0, nil
	}
	for i, f := range ro.bp.Fields {
		var (
			v Object
			m int64
		)
		v, m, err = readField(r, f.Spec)
		n += m
		if err != nil {
			return
		}
		ro.fields[i].Value = v
	}
	return
}

// Get returns the field's current value, or the field's zero value if the field is unset.
// Unknown field names return nil.
func (ro *RuntimeObject) Get(name string) Object {
	idx := ro.find(name)
	if idx < 0 {
		return nil
	}
	return ro.fields[idx].Value
}

// Set assigns a value to the named field after normalizing it per the field's Spec. Returns
// ErrFieldTypeMismatch if the value cannot be normalized or the field does not exist.
func (ro *RuntimeObject) Set(name string, v any) error {
	idx := ro.find(name)
	if idx < 0 {
		return fmt.Errorf("%w: unknown field %s", ErrFieldTypeMismatch, name)
	}

	spec := ro.bp.Fields[idx].Spec
	norm, err := normalize(spec, v)
	if err != nil {
		return fmt.Errorf("%w: field %s: %s", ErrFieldTypeMismatch, name, err)
	}

	ro.fields[idx].Value = norm
	return nil
}

func (ro *RuntimeObject) find(name string) int {
	for i, f := range ro.fields {
		if f.Name == name {
			return i
		}
	}
	return -1
}

// specZero returns the canonical zero value for a Spec.
func specZero(spec Object) Object {
	switch s := spec.(type) {
	case *PrimitiveSpec:
		return New(s.PrimitiveType.String())
	case *RefSpec:
		return New(s.Type.String())
	case *SliceSpec:
		sl, err := newRuntimeSlice(s.Type.String())
		if err != nil {
			// why: SliceSpec validation accepts any Type string; an unregistered element type
			// falls back to a heterogeneous slice so the RuntimeObject is still constructible.
			// The encode/decode path will surface the underlying error if the field is used.
			sl, _ = newRuntimeSlice("")
		}
		return sl
	case *MapSpec:
		rm, err := newRuntimeMap(s.KeyType.String(), s.ValueType.String())
		if err != nil {
			// why: MapSpec validation accepts any ValueType string; an unregistered element type
			// falls back to a heterogeneous map so the RuntimeObject is still constructible. The
			// encode/decode path will surface the underlying error if the field is used. Mirrors
			// the SliceSpec fallback above.
			rm, _ = newRuntimeMap(s.KeyType.String(), "")
		}
		return rm
	case *PtrSpec:
		return nil
	case *ObjectSpec:
		return nil
	}
	return nil
}

// writeField serializes a single field value according to its Spec.
func writeField(w io.Writer, spec Object, value Object) (int64, error) {
	switch spec.(type) {
	case *PrimitiveSpec, *RefSpec, *SliceSpec, *MapSpec:
		if value == nil {
			return 0, fmt.Errorf("nil value for spec %T", spec)
		}
		return value.WriteTo(w)

	case *PtrSpec:
		n, err := Bool(value != nil).WriteTo(w)
		if err != nil || value == nil {
			return n, err
		}
		m, err := value.WriteTo(w)
		return n + m, err

	case *ObjectSpec:
		if value == nil {
			return 0, fmt.Errorf("nil value for ObjectSpec")
		}
		return Encode(w, value)
	}
	return 0, fmt.Errorf("unknown spec %T", spec)
}

// readField reads a single field value according to its Spec.
func readField(r io.Reader, spec Object) (Object, int64, error) {
	switch s := spec.(type) {
	case *PrimitiveSpec:
		obj := New(s.PrimitiveType.String())
		if obj == nil {
			return nil, 0, fmt.Errorf("%w: %s", ErrBlueprintNotFound, s.PrimitiveType.String())
		}
		n, err := obj.ReadFrom(r)
		return obj, n, err

	case *RefSpec:
		obj := New(s.Type.String())
		if obj == nil {
			return nil, 0, fmt.Errorf("%w: %s", ErrBlueprintNotFound, s.Type.String())
		}
		n, err := obj.ReadFrom(r)
		return obj, n, err

	case *SliceSpec:
		rs, err := newRuntimeSlice(s.Type.String())
		if err != nil {
			return nil, 0, err
		}
		n, err := rs.ReadFrom(r)
		return rs, n, err
	case *MapSpec:
		rm, err := newRuntimeMap(s.KeyType.String(), s.ValueType.String())
		if err != nil {
			return nil, 0, err
		}
		n, err := rm.ReadFrom(r)
		return rm, n, err
	case *PtrSpec:
		var present Bool
		n, err := (&present).ReadFrom(r)
		if err != nil || !present {
			return nil, n, err
		}

		obj := New(s.Type.String())
		if obj == nil {
			return nil, n, fmt.Errorf("%w: %s", ErrBlueprintNotFound, s.Type.String())
		}
		m, err := obj.ReadFrom(r)
		return obj, n + m, err
	case *ObjectSpec:
		return Decode(r)
	}
	return nil, 0, fmt.Errorf("unknown spec %T", spec)
}
