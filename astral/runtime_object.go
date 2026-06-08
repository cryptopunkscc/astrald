package astral

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strings"
)

// isPtrNil reports whether a value carried in a PtrSpec slot represents "absent".
// Three equivalent forms exist:
//   - interface-nil (no Object stored)
//   - the canonical *Nil marker (what specZero/normalize produce)
//   - a typed nil pointer (e.g. (*Greeting)(nil) leaked from a caller)
//
// Mirrors ptrValue.IsNil() in ptr_value.go so the runtime Blueprint codec and the
// reflection-based Objectify codec agree on what "nil pointer" means on the wire.
func isPtrNil(value Object) bool {
	if value == nil {
		return true
	}
	if _, ok := value.(*Nil); ok {
		return true
	}
	v := reflect.ValueOf(value)
	return v.Kind() == reflect.Ptr && v.IsNil()
}

// MaxBlueprintDepth caps the nested RuntimeObject frames in a single encode/decode. A cycle
// closed by RefSpec or PtrSpec edges (e.g. mutually-recursive X↔Y Blueprints) would otherwise
// recurse unboundedly: RefSpec consumes zero bytes per frame, PtrSpec consumes one byte and
// recurses on every non-zero presence flag. The cap converts stack exhaustion into a typed
// ErrDepthExceeded.
const MaxBlueprintDepth = 64

// RuntimeObject is the live encoder/decoder for a runtime-registered Blueprint. Field values are
// kept in declared order alongside their normalized astral types.
type RuntimeObject struct {
	bp *Blueprint
	// struct branch (bp.Kind() == BlueprintKindStruct):
	fields []fieldValue
	index  map[string]int
	// alias branch (bp.Kind() == BlueprintKindAlias): value holds the underlying primitive
	// carrier (e.g. *Uint8). fields/index are nil. Wire bytes for an alias-typed value are
	// identical to the bytes of the underlying primitive; the alias name lives only in the
	// registry and surfaces through ObjectType.
	value Object
}

type fieldValue struct {
	Name  string
	Value Object
}

// NewRuntimeObject returns a RuntimeObject whose fields are initialized to the zero value of
// each Spec. Returns an error if the Blueprint fails validation — guards against duplicate
// Field.Name or other structural problems that would make field lookup ambiguous.
// Spec-zeros are resolved through defaultBlueprints; for a per-call registry, the codec
// path uses newRuntimeObjectWith internally.
func NewRuntimeObject(bp *Blueprint) (*RuntimeObject, error) {
	return newRuntimeObjectWith(defaultBlueprints, bp)
}

// newRuntimeObjectWith is the bps-aware constructor used by Blueprints.New when materializing
// a runtime Blueprint through a custom registry. Mirrors newRuntime{Slice,Array,Map}With.
// Dispatches on bp.Kind(): struct kind populates fields/index from bp.Fields; alias kind
// allocates the underlying primitive via bps.New(bp.Underlying).
func newRuntimeObjectWith(bps *Blueprints, bp *Blueprint) (*RuntimeObject, error) {
	return newRuntimeObjectAt(bps, bp, 0)
}

// newRuntimeObjectAt threads a construction-time depth counter so a RefSpec/PtrSpec cycle
// (A.RefSpec→B, B.RefSpec→A) recursing through specZero → bps.New → newRuntimeObjectAt
// surfaces ErrDepthExceeded instead of overflowing the Go stack. The cap mirrors the
// runtime depth wrapper at WriteTo/ReadFrom (MaxBlueprintDepth).
func newRuntimeObjectAt(bps *Blueprints, bp *Blueprint, depth int) (*RuntimeObject, error) {
	if bp == nil {
		return &RuntimeObject{}, nil
	}
	if depth > MaxBlueprintDepth {
		return nil, fmt.Errorf("%w: %s (construction)", ErrDepthExceeded, bp.Type)
	}
	if err := validateBlueprint(bp); err != nil {
		return nil, err
	}
	if bp.Kind() == BlueprintKindAlias {
		v := bps.New(bp.Underlying.String())
		if v == nil {
			return nil, fmt.Errorf("%w: %s", ErrBlueprintNotFound, bp.Underlying)
		}
		return &RuntimeObject{bp: bp, value: v}, nil
	}
	ro := &RuntimeObject{
		bp:     bp,
		fields: make([]fieldValue, 0, len(bp.Fields)),
		index:  make(map[string]int, len(bp.Fields)),
	}
	for i, f := range bp.Fields {
		name := f.Name.String()
		v, zerr := specZeroAtErr(bps, f.Spec, depth+1)
		if zerr != nil {
			return nil, zerr
		}
		ro.fields = append(ro.fields, fieldValue{Name: name, Value: v})
		ro.index[name] = i
	}
	return ro, nil
}

// Underlying returns the wrapped primitive Object for an alias-kind RuntimeObject, or nil
// for a struct-kind one. Callers may type-assert it to the specific astral primitive (e.g.
// *Uint8) named by the Blueprint's Underlying field.
func (ro *RuntimeObject) Underlying() Object {
	if ro.bp == nil || ro.bp.Kind() != BlueprintKindAlias {
		return nil
	}
	return ro.value
}

// GetRuntimeObject returns a fresh RuntimeObject backed by this Blueprint. Spec-zeros are
// resolved through defaultBlueprints; route through Blueprints.New for a custom registry.
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
	if ro.bp.Kind() == BlueprintKindAlias {
		return ro.value.WriteTo(w)
	}

	ow, ok := w.(*objectWriter)
	if !ok {
		ow = &objectWriter{Writer: w}
	}
	err = ow.enter(ro.bp.Type)
	defer ow.exit()
	if err != nil {
		return 0, err
	}

	for i, f := range ro.bp.Fields {
		m, ferr := writeField(ow, f.Spec, ro.fields[i].Value)
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
	if ro.bp.Kind() == BlueprintKindAlias {
		return ro.value.ReadFrom(r)
	}

	// why: inherit the wrapper's *Blueprints (set by Decode from cfg.Blueprints) so nested
	// PrimitiveSpec/RefSpec/PtrSpec resolutions use the caller's registry. A freshly
	// constructed wrapper has bps=nil, which or.resolve() maps to defaultBlueprints.
	or, ok := r.(*objectReader)
	if !ok {
		or = &objectReader{Reader: r}
	}
	err = or.enter(ro.bp.Type)
	defer or.exit()
	if err != nil {
		return 0, err
	}

	for i, f := range ro.bp.Fields {
		v, m, ferr := readField(or, f.Spec)
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

// specZero returns the canonical zero value for a Spec, resolving Primitive/Ref/Slice/Array/Map
// element types through the supplied registry so a custom Blueprints (WithBlueprints) sees its
// own types, not defaultBlueprints.
func specZero(bps *Blueprints, spec Spec) Object {
	return specZeroAt(bps, spec, 0)
}

// specZeroAt is the depth-aware variant. RefSpec materialization is the only branch that can
// recurse through newRuntimeObjectAt (a runtime Blueprint referring to another runtime
// Blueprint); a cycle is bounded by MaxBlueprintDepth. PrimitiveSpec resolves to a compile-time
// prototype and never recurses; container Specs construct empty carriers; PtrSpec/ObjectSpec
// return &Nil{}.
func specZeroAt(bps *Blueprints, spec Spec, depth int) Object {
	v, _ := specZeroAtErr(bps, spec, depth)
	return v
}

// specZeroAtErr is the error-propagating variant. Only ErrDepthExceeded escapes — other
// failures (unregistered type, etc.) still resolve to nil to preserve the documented
// "treat-as-unregistered" contract at the codec layer.
func specZeroAtErr(bps *Blueprints, spec Spec, depth int) (Object, error) {
	switch s := spec.(type) {
	case *PrimitiveSpec:
		return bps.New(s.PrimitiveType.String()), nil
	case *RefSpec:
		return newAt(bps, s.Type.String(), depth)
	case *SliceSpec:
		// why: no heterogeneous fallback when the element type is unregistered — would silently
		// swap wire shape and let encode succeed against bytes decode can't read. Codec layer
		// re-resolves and fails identically on both ends.
		sl, err := newRuntimeSliceWith(bps, s.Type.String())
		if err != nil {
			return nil, nil
		}
		return sl, nil
	case *ArraySpec:
		// why: same symmetry argument as SliceSpec.
		ra, err := newRuntimeArrayWith(bps, s.Type.String(), uint32(s.Length))
		if err != nil {
			return nil, nil
		}
		return ra, nil
	case *MapSpec:
		// why: no heterogeneous fallback for unregistered ValueType — mirrors SliceSpec above.
		rm, err := newRuntimeMapWith(bps, s.KeyType.String(), s.ValueType.String())
		if err != nil {
			return nil, nil
		}
		return rm, nil
	case *PtrSpec:
		// why: vision contract — Get returns a non-nil zero so callers can comma-ok type-assert
		// without nil-guards. *Nil is the canonical "absent" marker; writeField/normalize map it
		// to wire presence=0.
		return &Nil{}, nil
	case *ObjectSpec:
		// why: same vision contract. ObjectSpec's Nil tag round-trips cleanly via Encode/Decode.
		return &Nil{}, nil
	}
	return nil, nil
}

// writeField serializes a single field value according to its Spec.
func writeField(w io.Writer, spec Spec, value Object) (int64, error) {
	switch spec.(type) {
	case *PrimitiveSpec, *RefSpec:
		if value == nil {
			return 0, fmt.Errorf("nil value for spec %T", spec)
		}
		return value.WriteTo(w)

	case *SliceSpec:
		if value == nil {
			return 0, fmt.Errorf("nil value for SliceSpec")
		}
		return value.WriteTo(w)

	case *ArraySpec:
		if value == nil {
			return 0, fmt.Errorf("nil value for ArraySpec")
		}
		return value.WriteTo(w)

	case *MapSpec:
		if value == nil {
			return 0, fmt.Errorf("nil value for MapSpec")
		}
		return value.WriteTo(w)

	case *PtrSpec:
		// why: three equivalent "absent" forms — see isPtrNil. Mirrors the nil-flag protocol
		// implemented by ptrValue in ptr_value.go so reflection and runtime Blueprint codec
		// agree on the wire shape.
		absent := isPtrNil(value)
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

// MarshalJSON serializes the RuntimeObject. Struct kind emits a JSON object keyed by field
// name; alias kind emits the underlying primitive's JSON form directly (the alias name lives
// in ObjectType, not in the payload). A nil Blueprint marshals to null.
func (ro *RuntimeObject) MarshalJSON() ([]byte, error) {
	if ro.bp == nil {
		return jsonNull, nil
	}
	if ro.bp.Kind() == BlueprintKindAlias {
		return json.Marshal(ro.value)
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

// UnmarshalJSON populates a RuntimeObject from JSON. Struct kind reads a JSON object keyed by
// field name; missing keys leave the field at its spec-zero. Alias kind decodes the payload
// directly into the underlying primitive. The receiver must already be bound to a Blueprint
// (via NewRuntimeObject or astral.New(typeName)) — without it there's no schema to resolve
// field names against.
//
// Field-name matching is case-insensitive, mirroring encoding/json's default and the sibling
// structValue path (struct_value.go:145). Two payload keys differing only by case are
// rejected as ambiguous. Excess keys (no matching field) are silently ignored — same as
// encoding/json's default; use a schema-aware validator if drift detection is needed.
func (ro *RuntimeObject) UnmarshalJSON(data []byte) error {
	if ro.bp == nil {
		return errors.New("RuntimeObject.UnmarshalJSON: no Blueprint bound")
	}
	if ro.bp.Kind() == BlueprintKindAlias {
		// why: passing &ro.value (Object interface) to json.Unmarshal causes the JSON literal
		// `null` to nilify the interface itself, leaving the carrier without an underlying
		// primitive — subsequent WriteTo/MarshalJSON panic. Trim/equal-check the input and
		// dispatch into the held primitive instead.
		if bytes.Equal(bytes.TrimSpace(data), jsonNull) {
			return nil
		}
		if ro.value == nil {
			return errors.New("RuntimeObject.UnmarshalJSON: alias-kind has no underlying primitive bound")
		}
		return json.Unmarshal(data, ro.value)
	}
	var raw map[string]json.RawMessage
	err := json.Unmarshal(data, &raw)
	if err != nil {
		return err
	}

	// why: lowercase every incoming key for case-insensitive lookup. Two keys folding to
	// the same lowercase form is an ambiguous payload, not a silent winner-takes-last.
	for k, v := range raw {
		l := strings.ToLower(k)
		if k == l {
			continue
		}
		if _, dup := raw[l]; dup {
			return errors.New("object has duplicate fields due to case insensitivity")
		}
		raw[l] = v
		delete(raw, k)
	}

	for i, f := range ro.bp.Fields {
		rawField, ok := raw[strings.ToLower(f.Name.String())]
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
func marshalFieldJSON(spec Spec, value Object) (json.RawMessage, error) {
	if _, isPtr := spec.(*PtrSpec); isPtr && isPtrNil(value) {
		// why: PtrSpec uses null for absence to match the binary nil-flag protocol; this
		// keeps both encodings symmetric for the three "absent" forms isPtrNil recognises.
		return jsonNull, nil
	}
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
func unmarshalFieldJSON(spec Spec, raw json.RawMessage) (Object, error) {
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
//
// Name resolution for PrimitiveSpec/RefSpec/PtrSpec consults or.resolve() — the per-call
// registry threaded through Decode via WithBlueprints, falling back to defaultBlueprints
// when unset. NewRuntimeSlice/Array/Map still use the package-level resolver for element
// types; threading those is a separate follow-up (their nested element decode also
// resolves via package-level New).
func readField(or *objectReader, spec Spec) (Object, int64, error) {
	bps := or.resolve()
	switch s := spec.(type) {
	case *PrimitiveSpec:
		obj := bps.New(s.PrimitiveType.String())
		if obj == nil {
			return nil, 0, fmt.Errorf("%w: %s", ErrBlueprintNotFound, s.PrimitiveType.String())
		}
		n, err := obj.ReadFrom(or)
		return obj, n, err

	case *RefSpec:
		obj := bps.New(s.Type.String())
		if obj == nil {
			return nil, 0, fmt.Errorf("%w: %s", ErrBlueprintNotFound, s.Type.String())
		}
		n, err := obj.ReadFrom(or)
		return obj, n, err

	case *SliceSpec:
		rs, err := newRuntimeSliceWith(bps, s.Type.String())
		if err != nil {
			return nil, 0, err
		}
		n, err := rs.ReadFrom(or)
		return rs, n, err
	case *ArraySpec:
		ra, err := newRuntimeArrayWith(bps, s.Type.String(), uint32(s.Length))
		if err != nil {
			return nil, 0, err
		}
		n, err := ra.ReadFrom(or)
		return ra, n, err
	case *MapSpec:
		rm, err := newRuntimeMapWith(bps, s.KeyType.String(), s.ValueType.String())
		if err != nil {
			return nil, 0, err
		}
		n, err := rm.ReadFrom(or)
		return rm, n, err
	case *PtrSpec:
		// why: strict 0/1 presence flag — matches ptrValue.ReadFrom and
		// readRuntimeBlueprintPtr. A permissive any-non-zero accept lets a schema-divergent
		// peer or adversary feed arbitrary payload bytes that the next field then misreads.
		var presence Uint8
		n, err := (&presence).ReadFrom(or)
		if err != nil {
			return nil, n, err
		}
		switch presence {
		case 0:
			// canonical absent per the vision Get contract.
			return &Nil{}, n, nil
		case 1:
			// fall through to payload read
		default:
			return nil, n, fmt.Errorf("invalid PtrSpec presence flag: %d", presence)
		}

		obj := bps.New(s.Type.String())
		if obj == nil {
			return nil, n, fmt.Errorf("%w: %s", ErrBlueprintNotFound, s.Type.String())
		}
		m, err := obj.ReadFrom(or)
		return obj, n + m, err
	case *ObjectSpec:
		return Decode(or, WithBlueprints(bps))
	}
	return nil, 0, fmt.Errorf("unknown spec %T", spec)
}
