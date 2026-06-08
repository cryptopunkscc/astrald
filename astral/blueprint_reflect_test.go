package astral

import (
	"bytes"
	"errors"
	"io"
	"reflect"
	"testing"
)

// --- witness types used as field carriers across the dispatch tests ---

// reflectWitness implements Object with a value receiver so it lands in the
// "Object impl, not in primitive allowlist" branch of specFromType.
type reflectWitness struct{ X Uint32 }

// astral:blueprint-ignore
func (reflectWitness) ObjectType() string                    { return "test.reflect.witness" }
func (w reflectWitness) WriteTo(w2 io.Writer) (int64, error) { return Objectify(&w).WriteTo(w2) }
func (w *reflectWitness) ReadFrom(r io.Reader) (int64, error) {
	return Objectify(w).ReadFrom(r)
}

// ptrReceiverObject implements Object on *T only (not on T). Exercises the
// third probe path in tryObjectType.
type ptrReceiverObject struct{ X Uint32 }

// astral:blueprint-ignore
func (*ptrReceiverObject) ObjectType() string                   { return "test.reflect.ptr_recv" }
func (p *ptrReceiverObject) WriteTo(w io.Writer) (int64, error) { return Objectify(p).WriteTo(w) }
func (p *ptrReceiverObject) ReadFrom(r io.Reader) (int64, error) {
	return Objectify(p).ReadFrom(r)
}

// emptyTypeName implements Object but ObjectType() returns "". specFromType's
// Object probe must fall through to container dispatch instead of producing an
// empty-named Spec.
type emptyTypeName []byte

// astral:blueprint-ignore
func (emptyTypeName) ObjectType() string                   { return "" }
func (emptyTypeName) WriteTo(w io.Writer) (int64, error)   { return 0, nil }
func (*emptyTypeName) ReadFrom(r io.Reader) (int64, error) { return 0, nil }

// nonObjectStruct has no ObjectType — used to force errors in slice/map elem
// and the final fall-through branch.
type nonObjectStruct struct{ N int }

func mustReflectBlueprint(t *testing.T, v any) *Blueprint {
	t.Helper()
	bp, err := BlueprintFromType(reflect.TypeOf(v))
	if err != nil {
		t.Fatalf("BlueprintFromType: %v", err)
	}
	return bp
}

// --- §3 dispatch table ---

// primitiveBattery and leaves are declared at package scope because Go
// requires methods to be in the same package as the type — local types inside
// test functions cannot implement Object.
type primitiveBattery struct {
	S8  String8
	S16 String16
	S32 String32
	S64 String64
	U8  Uint8
	U16 Uint16
	U32 Uint32
	U64 Uint64
	B8  Bytes8
	B16 Bytes16
	B32 Bytes32
	B64 Bytes64
	Bo  Bool
	Tm  Time
	Id  Identity
	Oid ObjectID
	Nc  Nonce
	Dr  Duration
	Zn  Zone
}

// astral:blueprint-ignore
func (primitiveBattery) ObjectType() string                   { return "test.reflect.primitives" }
func (primitiveBattery) WriteTo(w io.Writer) (int64, error)   { return 0, nil }
func (*primitiveBattery) ReadFrom(r io.Reader) (int64, error) { return 0, nil }

// TestSpecFromType_Dispatch exhausts every branch of specFromType using
// single-field witness structs. The table is the dispatch matrix; a regression
// in branch ordering or container/leaf classification shows up here first.
func TestSpecFromType_Dispatch(t *testing.T) {
	bp, err := BlueprintFromType(reflect.TypeOf(primitiveBattery{}))
	if err != nil {
		t.Fatalf("primitiveBattery: %v", err)
	}

	wantPrimitives := []struct {
		field string
		name  string
	}{
		{"S8", "string8"}, {"S16", "string16"}, {"S32", "string32"}, {"S64", "string64"},
		{"U8", "uint8"}, {"U16", "uint16"}, {"U32", "uint32"}, {"U64", "uint64"},
		{"B8", "bytes8"}, {"B16", "bytes16"}, {"B32", "bytes32"}, {"B64", "bytes64"},
		{"Bo", "bool"}, {"Tm", "time"}, {"Id", "identity"},
		{"Oid", "object_id.sha256"},
		{"Nc", "nonce64"}, {"Dr", "duration"}, {"Zn", "zone"},
	}
	if len(bp.Fields) != len(wantPrimitives) {
		t.Fatalf("primitive count: want %d, got %d", len(wantPrimitives), len(bp.Fields))
	}
	for i, want := range wantPrimitives {
		if bp.Fields[i].Name.String() != want.field {
			t.Errorf("field[%d] name: want %s, got %s", i, want.field, bp.Fields[i].Name)
		}
		ps, ok := bp.Fields[i].Spec.(*PrimitiveSpec)
		if !ok {
			t.Errorf("field %s: want *PrimitiveSpec, got %T", want.field, bp.Fields[i].Spec)
			continue
		}
		if ps.PrimitiveType.String() != want.name {
			t.Errorf("field %s: want primitive %s, got %s", want.field, want.name, ps.PrimitiveType)
		}
	}

	// Each subtest exercises one dispatch branch with a single-field struct.
	t.Run("ObjectInterface", func(t *testing.T) {
		bp := mustReflectBlueprint(t, struct {
			reflectWitnessNamed
			F Object
		}{})
		// reflectWitnessNamed is embedded only to provide ObjectType() for the witness
		// struct; the field of interest is F.
		spec := lookupField(t, bp, "F")
		if _, ok := spec.(*ObjectSpec); !ok {
			t.Fatalf("want *ObjectSpec, got %T", spec)
		}
	})

	t.Run("PointerToObject", func(t *testing.T) {
		bp := mustReflectBlueprint(t, struct {
			reflectWitnessNamed
			F *Identity
		}{})
		spec := lookupField(t, bp, "F")
		ps, ok := spec.(*PtrSpec)
		if !ok || ps.Type.String() != "identity" {
			t.Fatalf("want *PtrSpec{identity}, got %#v", spec)
		}
	})

	t.Run("Ref_NonPrimitiveObject", func(t *testing.T) {
		bp := mustReflectBlueprint(t, struct {
			reflectWitnessNamed
			F Blueprint
		}{})
		spec := lookupField(t, bp, "F")
		rs, ok := spec.(*RefSpec)
		if !ok || rs.Type.String() != "astral.blueprint" {
			t.Fatalf("want *RefSpec{astral.blueprint}, got %#v", spec)
		}
	})

	t.Run("SliceOfObject_Heterogeneous", func(t *testing.T) {
		bp := mustReflectBlueprint(t, struct {
			reflectWitnessNamed
			F []Object
		}{})
		spec := lookupField(t, bp, "F")
		ss, ok := spec.(*SliceSpec)
		if !ok || ss.Type.String() != "" {
			t.Fatalf("want *SliceSpec{Type:\"\"}, got %#v", spec)
		}
	})

	t.Run("SliceOfPtrConcrete", func(t *testing.T) {
		bp := mustReflectBlueprint(t, struct {
			reflectWitnessNamed
			F []*Identity
		}{})
		spec := lookupField(t, bp, "F")
		ss, ok := spec.(*SliceSpec)
		if !ok || ss.Type.String() != "identity" {
			t.Fatalf("want *SliceSpec{identity}, got %#v", spec)
		}
	})

	t.Run("SliceOfLeafPrimitive", func(t *testing.T) {
		// Bytes32 backs []byte but the Object probe runs first — element resolves
		// as the primitive name, not a generic byte slice.
		bp := mustReflectBlueprint(t, struct {
			reflectWitnessNamed
			F []Bytes32
		}{})
		spec := lookupField(t, bp, "F")
		ss, ok := spec.(*SliceSpec)
		if !ok || ss.Type.String() != "bytes32" {
			t.Fatalf("want *SliceSpec{bytes32}, got %#v", spec)
		}
	})

	t.Run("ArrayOfObject", func(t *testing.T) {
		bp := mustReflectBlueprint(t, struct {
			reflectWitnessNamed
			F [3]*Identity
		}{})
		spec := lookupField(t, bp, "F")
		as, ok := spec.(*ArraySpec)
		if !ok || as.Type.String() != "identity" || as.Length != 3 {
			t.Fatalf("want *ArraySpec{identity,3}, got %#v", spec)
		}
	})

	t.Run("ArrayHeterogeneous", func(t *testing.T) {
		bp := mustReflectBlueprint(t, struct {
			reflectWitnessNamed
			F [2]Object
		}{})
		spec := lookupField(t, bp, "F")
		as, ok := spec.(*ArraySpec)
		if !ok || as.Type.String() != "" || as.Length != 2 {
			t.Fatalf("want *ArraySpec{\"\",2}, got %#v", spec)
		}
	})

	t.Run("ArrayZeroLength", func(t *testing.T) {
		bp := mustReflectBlueprint(t, struct {
			reflectWitnessNamed
			F [0]Identity
		}{})
		spec := lookupField(t, bp, "F")
		as, ok := spec.(*ArraySpec)
		if !ok || as.Length != 0 {
			t.Fatalf("want *ArraySpec{Length:0}, got %#v", spec)
		}
	})

	mapKeyCases := []struct {
		name    string
		ctor    func() any
		wantKey string
		wantVal string
	}{
		{"string", func() any {
			return struct {
				reflectWitnessNamed
				F map[string]*Identity
			}{}
		}, "string16", "identity"},
		{"uint8", func() any {
			return struct {
				reflectWitnessNamed
				F map[uint8]*Identity
			}{}
		}, "uint8", "identity"},
		{"uint16", func() any {
			return struct {
				reflectWitnessNamed
				F map[uint16]*Identity
			}{}
		}, "uint16", "identity"},
		{"uint32", func() any {
			return struct {
				reflectWitnessNamed
				F map[uint32]*Identity
			}{}
		}, "uint32", "identity"},
		{"uint64", func() any {
			return struct {
				reflectWitnessNamed
				F map[uint64]*Identity
			}{}
		}, "uint64", "identity"},
		{"typed_string_alias", func() any {
			return struct {
				reflectWitnessNamed
				F map[String16]*Identity
			}{}
		}, "string16", "identity"},
		{"heterogeneous_value", func() any {
			return struct {
				reflectWitnessNamed
				F map[string]Object
			}{}
		}, "string16", ""},
	}
	for _, tc := range mapKeyCases {
		t.Run("Map_"+tc.name, func(t *testing.T) {
			bp := mustReflectBlueprint(t, tc.ctor())
			spec := lookupField(t, bp, "F")
			ms, ok := spec.(*MapSpec)
			if !ok {
				t.Fatalf("want *MapSpec, got %T", spec)
			}
			if ms.KeyType.String() != tc.wantKey {
				t.Errorf("KeyType: want %s, got %s", tc.wantKey, ms.KeyType)
			}
			if ms.ValueType.String() != tc.wantVal {
				t.Errorf("ValueType: want %s, got %s", tc.wantVal, ms.ValueType)
			}
		})
	}

	// --- error branches ---

	errCases := []struct {
		name string
		ctor func() reflect.Type
	}{
		{"NonObjectInterface", func() reflect.Type {
			return reflect.TypeOf(struct {
				reflectWitnessNamed
				F io.Reader
			}{})
		}},
		{"PointerToNonObject", func() reflect.Type {
			return reflect.TypeOf(struct {
				reflectWitnessNamed
				F *int
			}{})
		}},
		{"SliceOfNonObject", func() reflect.Type {
			return reflect.TypeOf(struct {
				reflectWitnessNamed
				F []int
			}{})
		}},
		{"SliceOfRawByteSlice", func() reflect.Type {
			// [][]byte — outer Slice OK, inner []byte doesn't implement Object
			return reflect.TypeOf(struct {
				reflectWitnessNamed
				F [][]byte
			}{})
		}},
		{"SliceOfNonObjectInterface", func() reflect.Type {
			return reflect.TypeOf(struct {
				reflectWitnessNamed
				F []io.Reader
			}{})
		}},
		{"ArrayOfNonObject", func() reflect.Type {
			return reflect.TypeOf(struct {
				reflectWitnessNamed
				F [3]int
			}{})
		}},
		{"MapWithSignedKey", func() reflect.Type {
			return reflect.TypeOf(struct {
				reflectWitnessNamed
				F map[int]*Identity
			}{})
		}},
		{"MapWithPlatformUintKey", func() reflect.Type {
			// why: platform-width uint is rejected to keep mapKeyTypeName aligned with
			// supportedMapKey in map_value.go — a uint key would split content hashes
			// across architectures.
			return reflect.TypeOf(struct {
				reflectWitnessNamed
				F map[uint]*Identity
			}{})
		}},
		{"MapWithBoolKey", func() reflect.Type {
			return reflect.TypeOf(struct {
				reflectWitnessNamed
				F map[bool]*Identity
			}{})
		}},
		{"MapWithArrayKey", func() reflect.Type {
			return reflect.TypeOf(struct {
				reflectWitnessNamed
				F map[[8]byte]*Identity
			}{})
		}},
		{"MapWithNonObjectValue", func() reflect.Type {
			return reflect.TypeOf(struct {
				reflectWitnessNamed
				F map[string]int
			}{})
		}},
		{"PlainStructFieldNoObject", func() reflect.Type {
			return reflect.TypeOf(struct {
				reflectWitnessNamed
				F nonObjectStruct
			}{})
		}},
		{"UnsupportedKind_Chan", func() reflect.Type {
			return reflect.TypeOf(struct {
				reflectWitnessNamed
				F chan int
			}{})
		}},
		{"UnsupportedKind_Func", func() reflect.Type {
			return reflect.TypeOf(struct {
				reflectWitnessNamed
				F func()
			}{})
		}},
		{"UnsupportedKind_Float", func() reflect.Type {
			return reflect.TypeOf(struct {
				reflectWitnessNamed
				F float64
			}{})
		}},
		{"UnsupportedKind_SignedInt", func() reflect.Type {
			return reflect.TypeOf(struct {
				reflectWitnessNamed
				F int32
			}{})
		}},
		{"UnsupportedKind_PlatformInt", func() reflect.Type {
			return reflect.TypeOf(struct {
				reflectWitnessNamed
				F int
			}{})
		}},
		{"UnsupportedKind_PlatformUint", func() reflect.Type {
			return reflect.TypeOf(struct {
				reflectWitnessNamed
				F uint
			}{})
		}},
	}
	for _, tc := range errCases {
		t.Run("Err_"+tc.name, func(t *testing.T) {
			_, err := BlueprintFromType(tc.ctor())
			if err == nil {
				t.Fatalf("expected error, got none")
			}
		})
	}
}

// reflectWitnessNamed is embedded into anonymous structs to give them an
// ObjectType() so BlueprintFromType can succeed at the top level. The exported
// "ReflectWitnessNamed" field that embedding produces is filtered out by the
// per-test lookupField helper.
type reflectWitnessNamed struct{}

// astral:blueprint-ignore
func (reflectWitnessNamed) ObjectType() string                   { return "test.reflect.named" }
func (reflectWitnessNamed) WriteTo(w io.Writer) (int64, error)   { return 0, nil }
func (*reflectWitnessNamed) ReadFrom(r io.Reader) (int64, error) { return 0, nil }

// lookupField returns the Spec of the named field, t.Fatal-ing if absent.
func lookupField(t *testing.T, bp *Blueprint, name string) Object {
	t.Helper()
	for _, f := range bp.Fields {
		if f.Name.String() == name {
			return f.Spec
		}
	}
	t.Fatalf("field %s not found in blueprint", name)
	return nil
}

type leaves struct {
	ByteSlice Bytes32
	StrAlias  String16
	BoolAlias Bool
	IntAlias  Uint64
}

// astral:blueprint-ignore
func (leaves) ObjectType() string                   { return "test.reflect.leaves" }
func (leaves) WriteTo(w io.Writer) (int64, error)   { return 0, nil }
func (*leaves) ReadFrom(r io.Reader) (int64, error) { return 0, nil }

// TestSpecFromType_LeafPrimitivesBeatContainers is the load-bearing pin: the
// Object probe at specFromType:87 must run before the Slice/Map/Array Kind
// dispatch, otherwise astral primitives backed by container Go kinds (Bytes32
// = []byte, ObjectID = struct with [32]byte) silently degrade to container
// Specs and produce wire-incompatible Blueprints.
func TestSpecFromType_LeafPrimitivesBeatContainers(t *testing.T) {
	bp := mustReflectBlueprint(t, leaves{})
	wants := map[string]string{
		"ByteSlice": "bytes32",
		"StrAlias":  "string16",
		"BoolAlias": "bool",
		"IntAlias":  "uint64",
	}
	for _, f := range bp.Fields {
		ps, ok := f.Spec.(*PrimitiveSpec)
		if !ok {
			t.Errorf("%s: want *PrimitiveSpec, got %T", f.Name, f.Spec)
			continue
		}
		want := wants[f.Name.String()]
		if ps.PrimitiveType.String() != want {
			t.Errorf("%s: want %s, got %s", f.Name, want, ps.PrimitiveType)
		}
	}
}

// --- §2 top-level entry ---

func TestBlueprintFromType_PointerAndValueAreEquivalent(t *testing.T) {
	byVal, err := BlueprintFromType(reflect.TypeOf(reflectWitness{}))
	if err != nil {
		t.Fatalf("value: %v", err)
	}
	byPtr, err := BlueprintFromType(reflect.TypeOf(&reflectWitness{}))
	if err != nil {
		t.Fatalf("pointer: %v", err)
	}
	if byVal.Type != byPtr.Type || len(byVal.Fields) != len(byPtr.Fields) {
		t.Fatalf("value vs pointer entry differ: %#v vs %#v", byVal, byPtr)
	}
}

func TestBlueprintFromType_NonStructErrors(t *testing.T) {
	cases := []reflect.Type{
		reflect.TypeOf(0),
		reflect.TypeOf(""),
		reflect.TypeOf([]int{}),
		reflect.TypeOf(map[string]int{}),
	}
	for _, ty := range cases {
		t.Run(ty.String(), func(t *testing.T) {
			_, err := BlueprintFromType(ty)
			if err == nil {
				t.Fatalf("expected error")
			}
		})
	}
}

func TestBlueprintFromType_StructNotImplementingObject(t *testing.T) {
	_, err := BlueprintFromType(reflect.TypeOf(nonObjectStruct{}))
	if err == nil {
		t.Fatal("expected error for struct without ObjectType")
	}
}

// emptyObjectType implements Object but its ObjectType() returns "". The
// derivation must fail at the top level (validateBlueprint rejects empty Type).
type emptyObjectType struct{ X Uint32 }

// astral:blueprint-ignore
func (emptyObjectType) ObjectType() string                   { return "" }
func (emptyObjectType) WriteTo(w io.Writer) (int64, error)   { return 0, nil }
func (*emptyObjectType) ReadFrom(r io.Reader) (int64, error) { return 0, nil }

func TestBlueprintFromType_EmptyObjectTypeRejected(t *testing.T) {
	_, err := BlueprintFromType(reflect.TypeOf(emptyObjectType{}))
	if err == nil {
		t.Fatal("expected error for empty ObjectType")
	}
}

type unexportedFieldHolder struct {
	X Uint32
	y Uint32 // skipped
}

// astral:blueprint-ignore
func (unexportedFieldHolder) ObjectType() string                   { return "test.reflect.unexported" }
func (unexportedFieldHolder) WriteTo(w io.Writer) (int64, error)   { return 0, nil }
func (*unexportedFieldHolder) ReadFrom(r io.Reader) (int64, error) { return 0, nil }

func TestBlueprintFromType_UnexportedFieldsSkipped(t *testing.T) {
	bp := mustReflectBlueprint(t, unexportedFieldHolder{})
	if len(bp.Fields) != 1 {
		t.Fatalf("want 1 exported field, got %d", len(bp.Fields))
	}
	if bp.Fields[0].Name.String() != "X" {
		t.Fatalf("want field X, got %s", bp.Fields[0].Name)
	}
}

type orderedFields struct {
	C Uint32
	A Uint32
	B Uint32
}

// astral:blueprint-ignore
func (orderedFields) ObjectType() string                   { return "test.reflect.order" }
func (orderedFields) WriteTo(w io.Writer) (int64, error)   { return 0, nil }
func (*orderedFields) ReadFrom(r io.Reader) (int64, error) { return 0, nil }

// Field ordering is the schema — the vision contract says reordering produces
// a different Blueprint. Lock in that derivation preserves declaration order.
func TestBlueprintFromType_FieldOrderPreserved(t *testing.T) {
	bp := mustReflectBlueprint(t, orderedFields{})
	want := []string{"C", "A", "B"}
	for i, name := range want {
		if bp.Fields[i].Name.String() != name {
			t.Fatalf("field[%d]: want %s, got %s", i, name, bp.Fields[i].Name)
		}
	}
}

func TestBlueprintOf_MatchesBlueprintFromType(t *testing.T) {
	v := reflectWitness{}
	a, err := BlueprintOf(&v)
	if err != nil {
		t.Fatal(err)
	}
	b, err := BlueprintFromType(reflect.TypeOf(&v))
	if err != nil {
		t.Fatal(err)
	}

	idA, err := ResolveObjectID(a)
	if err != nil {
		t.Fatal(err)
	}
	idB, err := ResolveObjectID(b)
	if err != nil {
		t.Fatal(err)
	}
	if idA.String() != idB.String() {
		t.Fatalf("expected identical IDs, got %s vs %s", idA, idB)
	}
}

func TestBlueprintFromType_ErrorWrapping(t *testing.T) {
	type bad struct{ F int32 }
	// bad doesn't implement Object, so the top-level concreteObjectTypeOf errors
	// before any field walk — confirms the outer wrapper at :29-31.
	_, err := BlueprintFromType(reflect.TypeOf(bad{}))
	if err == nil {
		t.Fatal("expected error")
	}
}

// --- §6 behavior pins ---

type embeddedHolder struct {
	Identity // embedded — surfaces as field "Identity"
	X        Uint32
}

// astral:blueprint-ignore
func (embeddedHolder) ObjectType() string                   { return "test.reflect.embed" }
func (embeddedHolder) WriteTo(w io.Writer) (int64, error)   { return 0, nil }
func (*embeddedHolder) ReadFrom(r io.Reader) (int64, error) { return 0, nil }

// Embedded fields are not special-cased; their reflect.StructField.Name is the
// embedded type's name. Lock this in so a future "auto-flatten embedded"
// refactor is a deliberate decision.
func TestBlueprintFromType_EmbeddedFieldSurfacesByTypeName(t *testing.T) {
	bp := mustReflectBlueprint(t, embeddedHolder{})
	if len(bp.Fields) != 2 {
		t.Fatalf("want 2 fields, got %d", len(bp.Fields))
	}
	if bp.Fields[0].Name.String() != "Identity" {
		t.Fatalf("first field: want Identity, got %s", bp.Fields[0].Name)
	}
	// The embedded Identity is a value field — dispatch sees the struct, Object
	// probe succeeds, primitive allowlist hit → PrimitiveSpec{identity}.
	ps, ok := bp.Fields[0].Spec.(*PrimitiveSpec)
	if !ok || ps.PrimitiveType.String() != "identity" {
		t.Fatalf("embedded Identity: want PrimitiveSpec{identity}, got %#v", bp.Fields[0].Spec)
	}
}

type selfRefHolder struct {
	Next *selfRefHolder
}

// astral:blueprint-ignore
func (selfRefHolder) ObjectType() string                   { return "test.reflect.selfref" }
func (selfRefHolder) WriteTo(w io.Writer) (int64, error)   { return 0, nil }
func (*selfRefHolder) ReadFrom(r io.Reader) (int64, error) { return 0, nil }

// validateBlueprint rejects self-PtrSpec to bound decode recursion. Reflection
// produces such a spec from a recursive Go type; confirm the validator catches
// it and the error type matches.
func TestBlueprintFromType_SelfReferentialPtrRejected(t *testing.T) {
	_, err := BlueprintFromType(reflect.TypeOf(selfRefHolder{}))
	if !errors.Is(err, ErrBlueprintInvalid) {
		t.Fatalf("want ErrBlueprintInvalid, got %v", err)
	}
}

type doublePtrHolder struct {
	F **Identity
}

// astral:blueprint-ignore
func (doublePtrHolder) ObjectType() string                   { return "test.reflect.dblptr" }
func (doublePtrHolder) WriteTo(w io.Writer) (int64, error)   { return 0, nil }
func (*doublePtrHolder) ReadFrom(r io.Reader) (int64, error) { return 0, nil }

// **T silently flattens to *PtrSpec{T.ObjectType()}: the outer Ptr branch
// calls concreteObjectTypeOf(t.Elem()), which for *Identity hits the pointer
// path inside concreteObjectTypeOf and returns "identity". The outer Ptr
// level is dropped — no nested-nilability semantics. Pin the actual behavior;
// changing it would be a deliberate design call.
func TestBlueprintFromType_DoublePointerFlattens(t *testing.T) {
	bp, err := BlueprintFromType(reflect.TypeOf(doublePtrHolder{}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	spec := lookupField(t, bp, "F")
	ps, ok := spec.(*PtrSpec)
	if !ok || ps.Type.String() != "identity" {
		t.Fatalf("want *PtrSpec{identity} (flattened), got %#v", spec)
	}
}

// BlueprintOf(nil) is reachable through user code (passing an untyped nil
// Object). reflect.TypeOf(nil) returns nil; pin the behavior so a future
// regression is caught.
func TestBlueprintOf_NilType(t *testing.T) {
	defer func() {
		// Either panic or error is acceptable; the test fails only if the call
		// silently returns a non-nil Blueprint.
		_ = recover()
	}()
	bp, err := BlueprintOf(nil)
	if err == nil && bp != nil {
		t.Fatalf("BlueprintOf(nil) silently returned %#v", bp)
	}
}

// tryObjectType's empty-name guards are not surfaced through any public error;
// instead, they let the dispatch fall through to the container branch. The
// emptyTypeName witness implements Object with ObjectType()=="" and is backed
// by []byte. The probe must yield ("", false), the dispatch must then enter
// the Slice branch, and the slice's element (byte) won't implement Object —
// erroring on the element. This roundabout assertion is the only way to pin
// the guard from the public API.
func TestSpecFromType_EmptyObjectNameFallsThroughToContainer(t *testing.T) {
	_, err := BlueprintFromType(reflect.TypeOf(struct {
		reflectWitnessNamed
		F emptyTypeName
	}{}))
	if err == nil {
		t.Fatal("expected error — emptyTypeName must fall through to []byte container, which errors on byte elem")
	}
}

// ptrReceiverObject implements Object on *T only. Pin that tryObjectType's
// third probe (reflect.New(t).Interface()) catches it — otherwise a field of
// type ptrReceiverObject (value, not pointer) would error.
type ptrReceiverHolder struct {
	F ptrReceiverObject
}

// astral:blueprint-ignore
func (ptrReceiverHolder) ObjectType() string                   { return "test.reflect.ptr_recv_holder" }
func (ptrReceiverHolder) WriteTo(w io.Writer) (int64, error)   { return 0, nil }
func (*ptrReceiverHolder) ReadFrom(r io.Reader) (int64, error) { return 0, nil }

func TestSpecFromType_PointerReceiverObjectResolves(t *testing.T) {
	bp := mustReflectBlueprint(t, ptrReceiverHolder{})
	spec := lookupField(t, bp, "F")
	rs, ok := spec.(*RefSpec)
	if !ok || rs.Type.String() != "test.reflect.ptr_recv" {
		t.Fatalf("want *RefSpec{test.reflect.ptr_recv}, got %#v", spec)
	}
}

// --- §5 integration ---

type integrationStruct struct {
	N Uint32
	S String16
	P *Identity
}

// astral:blueprint-ignore
func (integrationStruct) ObjectType() string                   { return "test.reflect.integration" }
func (integrationStruct) WriteTo(w io.Writer) (int64, error)   { return 0, nil }
func (*integrationStruct) ReadFrom(r io.Reader) (int64, error) { return 0, nil }

// Reflection-derived Blueprints must be deterministic across calls — two
// derivations from the same Go type must produce identical ObjectIDs, which
// is what makes Blueprints content-addressable across peers.
func TestBlueprintFromType_DeterministicObjectID(t *testing.T) {
	a, err := BlueprintFromType(reflect.TypeOf(integrationStruct{}))
	if err != nil {
		t.Fatal(err)
	}
	b, err := BlueprintFromType(reflect.TypeOf(integrationStruct{}))
	if err != nil {
		t.Fatal(err)
	}
	idA, err := ResolveObjectID(a)
	if err != nil {
		t.Fatal(err)
	}
	idB, err := ResolveObjectID(b)
	if err != nil {
		t.Fatal(err)
	}
	if idA.String() != idB.String() {
		t.Fatalf("non-deterministic derivation: %s vs %s", idA, idB)
	}
}

// Encode/decode round-trip on the derived Blueprint confirms the result is
// itself a valid Object on the wire — the basic Object contract.
func TestBlueprintFromType_RoundTrip(t *testing.T) {
	src, err := BlueprintFromType(reflect.TypeOf(integrationStruct{}))
	if err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	if _, err := src.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}

	dst := &Blueprint{}
	if _, err := dst.ReadFrom(&buf); err != nil {
		t.Fatal(err)
	}

	if dst.Type != src.Type {
		t.Fatalf("type: want %s, got %s", src.Type, dst.Type)
	}
	if len(dst.Fields) != len(src.Fields) {
		t.Fatalf("field count: want %d, got %d", len(src.Fields), len(dst.Fields))
	}
}

type runtimeDriveable struct {
	N Uint32
	S String16
}

// astral:blueprint-ignore
func (runtimeDriveable) ObjectType() string                   { return "test.reflect.drive" }
func (runtimeDriveable) WriteTo(w io.Writer) (int64, error)   { return 0, nil }
func (*runtimeDriveable) ReadFrom(r io.Reader) (int64, error) { return 0, nil }

// End-to-end: derive a Blueprint, register it on an isolated registry, build
// a RuntimeObject from it, populate it, encode/decode, read fields back. This
// is the integration that proves derivation produces something the rest of
// the system can drive.
func TestBlueprintFromType_DrivesRuntimeObject(t *testing.T) {
	bp, err := BlueprintFromType(reflect.TypeOf(runtimeDriveable{}))
	if err != nil {
		t.Fatal(err)
	}

	ro, err := NewRuntimeObject(bp)
	if err != nil {
		t.Fatal(err)
	}
	if err := ro.Set("N", uint32(99)); err != nil {
		t.Fatal(err)
	}
	if err := ro.Set("S", "ok"); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	if _, err := ro.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}

	dst, err := NewRuntimeObject(bp)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := dst.ReadFrom(&buf); err != nil {
		t.Fatal(err)
	}

	if u, ok := dst.Get("N").(*Uint32); !ok || *u != 99 {
		t.Fatalf("N: want 99, got %#v", dst.Get("N"))
	}
	if s, ok := dst.Get("S").(*String16); !ok || *s != "ok" {
		t.Fatalf("S: want \"ok\", got %#v", dst.Get("S"))
	}
}
