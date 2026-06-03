package astral

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

// MaxBlueprintDepth caps the nested RuntimeObject frames in a single encode/decode. A cycle
// closed by RefSpec or PtrSpec edges (e.g. mutually-recursive X↔Y Blueprints) would otherwise
// recurse unboundedly: RefSpec consumes zero bytes per frame, PtrSpec consumes one byte and
// recurses on every non-zero presence flag. The cap converts stack exhaustion into a typed
// ErrDepthExceeded.
const MaxBlueprintDepth = 64

// RuntimeObject is the live encoder/decoder for a runtime-registered Blueprint. Field values are
// kept in declared order alongside their normalized astral types.
type RuntimeObject struct {
	bp     *Blueprint
	fields []fieldValue
	index  map[string]int
}

type fieldValue struct {
	Name  string
	Value Object
}

// NewRuntimeObject returns a RuntimeObject whose fields are initialized to the zero value of
// each Spec. Returns an error if the Blueprint fails validation — guards against duplicate
// Field.Name or other structural problems that would make field lookup ambiguous.
func NewRuntimeObject(bp *Blueprint) (*RuntimeObject, error) {
	if bp == nil {
		return &RuntimeObject{}, nil
	}
	if err := validateBlueprint(bp); err != nil {
		return nil, err
	}
	ro := &RuntimeObject{
		bp:     bp,
		fields: make([]fieldValue, 0, len(bp.Fields)),
		index:  make(map[string]int, len(bp.Fields)),
	}
	for i, f := range bp.Fields {
		name := f.Name.String()
		ro.fields = append(ro.fields, fieldValue{Name: name, Value: specZero(f.Spec)})
		ro.index[name] = i
	}
	return ro, nil
}

// GetRuntimeObject returns a fresh RuntimeObject backed by this Blueprint.
func (bp *Blueprint) GetRuntimeObject() (*RuntimeObject, error) { return NewRuntimeObject(bp) }

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

	dw, ok := w.(*depthWriter)
	if !ok {
		dw = &depthWriter{Writer: w}
	}
	err = dw.enter(ro.bp.Type)
	defer dw.exit()
	if err != nil {
		return 0, err
	}

	for i, f := range ro.bp.Fields {
		m, ferr := writeField(dw, f.Spec, ro.fields[i].Value)
		n += m
		if ferr != nil {
			return n, fmt.Errorf("%s: field %q: %w", ro.bp.Type, f.Name, ferr)
		}
	}
	return
}

func (ro *RuntimeObject) ReadFrom(r io.Reader) (n int64, err error) {
	if ro.bp == nil {
		return 0, nil
	}

	dr, ok := r.(*depthReader)
	if !ok {
		dr = &depthReader{Reader: r}
	}
	err = dr.enter(ro.bp.Type)
	defer dr.exit()
	if err != nil {
		return 0, err
	}

	for i, f := range ro.bp.Fields {
		v, m, ferr := readField(dr, f.Spec)
		n += m
		if ferr != nil {
			return n, fmt.Errorf("%s: field %q: %w", ro.bp.Type, f.Name, ferr)
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
	if i, ok := ro.index[name]; ok {
		return i
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
		// why: no heterogeneous fallback when the element type is unregistered — would silently
		// swap wire shape and let encode succeed against bytes decode can't read. Codec layer
		// re-resolves and fails identically on both ends.
		sl, err := NewRuntimeSlice(s.Type.String())
		if err != nil {
			return nil
		}
		return sl
	case *ArraySpec:
		// why: same symmetry argument as SliceSpec.
		ra, err := NewRuntimeArray(s.Type.String(), uint32(s.Length))
		if err != nil {
			return nil
		}
		return ra
	case *MapSpec:
		// why: no heterogeneous fallback for unregistered ValueType — mirrors SliceSpec above.
		rm, err := NewRuntimeMap(s.KeyType.String(), s.ValueType.String())
		if err != nil {
			return nil
		}
		return rm
	case *PtrSpec:
		// why: vision contract — Get returns a non-nil zero so callers can comma-ok type-assert
		// without nil-guards. *Nil is the canonical "absent" marker; writeField/normalize map it
		// to wire presence=0.
		return &Nil{}
	case *ObjectSpec:
		// why: same vision contract. ObjectSpec's Nil tag round-trips cleanly via Encode/Decode.
		return &Nil{}
	}
	return nil
}

// writeField serializes a single field value according to its Spec.
func writeField(w io.Writer, spec Object, value Object) (int64, error) {
	switch s := spec.(type) {
	case *PrimitiveSpec, *RefSpec:
		if value == nil {
			return 0, fmt.Errorf("nil value for spec %T", spec)
		}
		return value.WriteTo(w)

	case *SliceSpec:
		// why: re-resolve so encode fails with the same ErrBlueprintNotFound that decode
		// would — keeps the codec symmetric when the element type is unregistered.
		if _, err := NewRuntimeSlice(s.Type.String()); err != nil {
			return 0, err
		}
		if value == nil {
			return 0, fmt.Errorf("nil value for SliceSpec")
		}
		return value.WriteTo(w)

	case *ArraySpec:
		// why: same symmetry argument as SliceSpec above.
		if _, err := NewRuntimeArray(s.Type.String(), uint32(s.Length)); err != nil {
			return 0, err
		}
		if value == nil {
			return 0, fmt.Errorf("nil value for ArraySpec")
		}
		return value.WriteTo(w)

	case *MapSpec:
		// why: same symmetry argument as SliceSpec above.
		if _, err := NewRuntimeMap(s.KeyType.String(), s.ValueType.String()); err != nil {
			return 0, err
		}
		if value == nil {
			return 0, fmt.Errorf("nil value for MapSpec")
		}
		return value.WriteTo(w)

	case *PtrSpec:
		// why: *Nil is the canonical "absent" carrier (specZero returns it; normalize maps
		// untyped nil to it). Detect it here and emit presence=0 + no payload.
		_, isNil := value.(*Nil)
		absent := value == nil || isNil
		n, err := Bool(!absent).WriteTo(w)
		if err != nil || absent {
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

// MarshalJSON serializes the RuntimeObject as a JSON object keyed by field name. The walk
// mirrors WriteTo's field order; each value's JSON shape is determined by its Spec. A nil
// Blueprint marshals to null.
func (ro *RuntimeObject) MarshalJSON() ([]byte, error) {
	if ro.bp == nil {
		return jsonNull, nil
	}
	out := make(map[string]json.RawMessage, len(ro.bp.Fields))
	for i, f := range ro.bp.Fields {
		raw, err := marshalFieldJSON(f.Spec, ro.fields[i].Value)
		if err != nil {
			return nil, fmt.Errorf("%s: field %q: %w", ro.bp.Type, f.Name, err)
		}
		out[f.Name.String()] = raw
	}
	return json.Marshal(out)
}

// UnmarshalJSON populates fields from a JSON object keyed by field name. Missing keys leave
// the field at its spec-zero from NewRuntimeObject. The receiver must already be bound to a
// Blueprint (via NewRuntimeObject or astral.New(typeName)) — without it there's no schema to
// resolve field names against.
func (ro *RuntimeObject) UnmarshalJSON(data []byte) error {
	if ro.bp == nil {
		return errors.New("RuntimeObject.UnmarshalJSON: no Blueprint bound")
	}
	var raw map[string]json.RawMessage
	err := json.Unmarshal(data, &raw)
	if err != nil {
		return err
	}
	for i, f := range ro.bp.Fields {
		rawField, ok := raw[f.Name.String()]
		if !ok {
			// why: absent on the wire — leave the spec-zero NewRuntimeObject installed.
			continue
		}
		v, err := unmarshalFieldJSON(f.Spec, rawField)
		if err != nil {
			return fmt.Errorf("%s: field %q: %w", ro.bp.Type, f.Name, err)
		}
		ro.fields[i].Value = v
	}
	return nil
}

// marshalFieldJSON serializes a single field value per its Spec. ObjectSpec wraps the payload
// in JSONAdapter so the receiver can resolve the concrete type from the wire — the schema
// alone doesn't pin it.
func marshalFieldJSON(spec Object, value Object) (json.RawMessage, error) {
	if value == nil {
		return nil, fmt.Errorf("nil value for spec %T", spec)
	}
	switch spec.(type) {
	case *PrimitiveSpec, *RefSpec, *SliceSpec, *ArraySpec, *MapSpec, *PtrSpec:
		return json.Marshal(value)
	case *ObjectSpec:
		// why: *Nil is the canonical absent carrier; emit a bare null instead of an envelope
		// with Type:"nil" so the wire matches the PtrSpec convention and the receiver's null
		// fast-path handles both Specs identically.
		if _, isNil := value.(*Nil); isNil {
			return jsonNull, nil
		}
		inner, err := json.Marshal(value)
		if err != nil {
			return nil, err
		}
		return json.Marshal(JSONAdapter{
			Type:   value.ObjectType(),
			Object: inner,
		})
	}
	return nil, fmt.Errorf("unknown spec %T", spec)
}

// unmarshalFieldJSON decodes a single field's JSON payload per its Spec. Each branch
// instantiates the canonical wire type, unmarshals into it, and returns the populated value.
// null on a PtrSpec or ObjectSpec resolves to the canonical &Nil{} carrier.
func unmarshalFieldJSON(spec Object, raw json.RawMessage) (Object, error) {
	switch s := spec.(type) {
	case *PrimitiveSpec:
		obj := New(s.PrimitiveType.String())
		if obj == nil {
			return nil, fmt.Errorf("%w: %s", ErrBlueprintNotFound, s.PrimitiveType)
		}
		err := json.Unmarshal(raw, &obj)
		if err != nil {
			return nil, err
		}
		return obj, nil

	case *RefSpec:
		obj := New(s.Type.String())
		if obj == nil {
			return nil, fmt.Errorf("%w: %s", ErrBlueprintNotFound, s.Type)
		}
		err := json.Unmarshal(raw, &obj)
		if err != nil {
			return nil, err
		}
		return obj, nil

	case *SliceSpec:
		rs, err := NewRuntimeSlice(s.Type.String())
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(raw, rs)
		if err != nil {
			return nil, err
		}
		return rs, nil

	case *ArraySpec:
		ra, err := NewRuntimeArray(s.Type.String(), uint32(s.Length))
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(raw, ra)
		if err != nil {
			return nil, err
		}
		return ra, nil

	case *MapSpec:
		rm, err := NewRuntimeMap(s.KeyType.String(), s.ValueType.String())
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(raw, rm)
		if err != nil {
			return nil, err
		}
		return rm, nil

	case *PtrSpec:
		if bytes.Equal(bytes.TrimSpace(raw), jsonNull) {
			return &Nil{}, nil
		}
		obj := New(s.Type.String())
		if obj == nil {
			return nil, fmt.Errorf("%w: %s", ErrBlueprintNotFound, s.Type)
		}
		err := json.Unmarshal(raw, &obj)
		if err != nil {
			return nil, err
		}
		return obj, nil

	case *ObjectSpec:
		if bytes.Equal(bytes.TrimSpace(raw), jsonNull) {
			return &Nil{}, nil
		}
		var env JSONAdapter
		err := json.Unmarshal(raw, &env)
		if err != nil {
			return nil, err
		}
		obj := New(env.Type)
		if obj == nil {
			return nil, fmt.Errorf("%w: %s", ErrBlueprintNotFound, env.Type)
		}
		if env.Object != nil {
			err = json.Unmarshal(env.Object, &obj)
			if err != nil {
				return nil, err
			}
		}
		return obj, nil
	}
	return nil, fmt.Errorf("unknown spec %T", spec)
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
		rs, err := NewRuntimeSlice(s.Type.String())
		if err != nil {
			return nil, 0, err
		}
		n, err := rs.ReadFrom(r)
		return rs, n, err
	case *ArraySpec:
		ra, err := NewRuntimeArray(s.Type.String(), uint32(s.Length))
		if err != nil {
			return nil, 0, err
		}
		n, err := ra.ReadFrom(r)
		return ra, n, err
	case *MapSpec:
		rm, err := NewRuntimeMap(s.KeyType.String(), s.ValueType.String())
		if err != nil {
			return nil, 0, err
		}
		n, err := rm.ReadFrom(r)
		return rm, n, err
	case *PtrSpec:
		// note: presence byte is read permissively (any non-zero → present). Under schema
		// divergence (peer wrote a different field shape) the byte here may be arbitrary
		// payload; consider requiring strict 0/1 to fail fast instead of recursing on
		// the remaining stream.
		var present Bool
		n, err := (&present).ReadFrom(r)
		if err != nil {
			return nil, n, err
		}
		if !present {
			// why: canonical absent representation per the vision Get contract.
			return &Nil{}, n, nil
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
