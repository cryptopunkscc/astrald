package astral_test

import (
	"encoding/json"
	"io"
	"testing"

	"github.com/cryptopunkscc/astrald/astral"
)

// --- test-only custom object A ---
type TestObjectA struct {
	Value astral.String
}

func (t TestObjectA) ObjectType() string { return "test.object.a" }

func (t TestObjectA) WriteTo(w io.Writer) (n int64, err error) {
	return astral.Struct(t).WriteTo(w)
}

func (t *TestObjectA) ReadFrom(r io.Reader) (n int64, err error) {
	return astral.Struct(t).ReadFrom(r)
}

// --- test-only custom object B ---
type TestObjectB struct {
	Number astral.Int8
}

func (t TestObjectB) ObjectType() string { return "test.object.b" }

func (t TestObjectB) WriteTo(w io.Writer) (n int64, err error) {
	return astral.Struct(t).WriteTo(w)
}

func (t *TestObjectB) ReadFrom(r io.Reader) (n int64, err error) {
	return astral.Struct(t).ReadFrom(r)
}

func P[T any](v T) *T {
	return &v
}

func TestSlice_MarshalJSON_Empty(t *testing.T) {
	elems := []astral.Object{}
	s := astral.WrapSlice(&elems)

	b, err := s.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON returned error: %v", err)
	}

	var out []any
	if err := json.Unmarshal(b, &out); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if len(out) != 0 {
		t.Fatalf("expected empty array, got %v", out)
	}
}

func TestSlice_MarshalJSON_CustomMixed(t *testing.T) {
	astral.DefaultBlueprints.Add(&TestObjectA{})
	astral.DefaultBlueprints.Add(&TestObjectB{})

	objects := []astral.Object{
		&TestObjectA{Value: "hello"},
		&TestObjectB{Number: 99},
	}

	slice := astral.WrapSlice(&objects)

	data, err := slice.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON returned error: %v", err)
	}

	// Decode encoded objects → adapters
	var adapters []astral.JSONEncodeAdapter
	if err := json.Unmarshal(data, &adapters); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if len(adapters) != 2 {
		t.Fatalf("expected 2 items, got %d", len(adapters))
	}

	for i, adapter := range adapters {
		switch orig := objects[i].(type) {

		case *TestObjectA:
			if adapter.Type != "test.object.a" {
				t.Fatalf("expected type test.object.a, got %s", adapter.Type)
			}

			// adapter.Object must be a *decoded object*, not same pointer
			objBytes, _ := json.Marshal(adapter.Object)
			if string(objBytes) == "" {
				t.Fatalf("expected Object field, got empty")
			}

			// decode into new struct to compare values
			var decoded TestObjectA
			decodedBytes, _ := json.Marshal(adapter.Object)
			_ = json.Unmarshal(decodedBytes, &decoded)

			if decoded.Value != orig.Value {
				t.Fatalf("expected Value=%q, got %q", orig.Value, decoded.Value)
			}

			if adapter.Object == orig {
				t.Fatalf("expected new JSON-mapped object, got original pointer")
			}

		case *TestObjectB:
			if adapter.Type != "test.object.b" {
				t.Fatalf("expected type test.object.b, got %s", adapter.Type)
			}

			objBytes, _ := json.Marshal(adapter.Object)
			if string(objBytes) == "" {
				t.Fatalf("expected Object field, got empty")
			}

			var decoded TestObjectB
			decodedBytes, _ := json.Marshal(adapter.Object)
			_ = json.Unmarshal(decodedBytes, &decoded)

			if decoded.Number != orig.Number {
				t.Fatalf("expected Number=%d, got %d", orig.Number, decoded.Number)
			}

			if adapter.Object == orig {
				t.Fatalf("expected new JSON-mapped object, got original pointer")
			}

		default:
			t.Fatalf("unexpected object type at index %d: %T", i, objects[i])
		}
	}

	var decodedSlice astral.Slice[astral.Object]
	if err := decodedSlice.UnmarshalJSON(data); err != nil {
		t.Fatalf("round-trip Unmarshal failed: %v", err)
	}

	if len(*decodedSlice.Elem) != 2 {
		t.Fatalf("round-trip len mismatch: expected 2, got %d", len(*decodedSlice.Elem))
	}

	// deep semantic equality
	if (*decodedSlice.Elem)[0].(*TestObjectA).Value != "hello" ||
		(*decodedSlice.Elem)[1].(*TestObjectB).Number != 99 {
		t.Fatalf("round-trip values mismatch: %#v", *decodedSlice.Elem)
	}
}

func TestSlice_MarshalJSON_Int8Only(t *testing.T) {
	objects := []astral.Object{
		P(astral.Int8(11)),
		P(astral.Int8(99)),
	}

	slice := astral.WrapSlice(&objects)

	data, err := slice.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON returned error: %v", err)
	}

	// Decode encoded objects → adapters
	var adapters []astral.JSONEncodeAdapter
	if err := json.Unmarshal(data, &adapters); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if len(adapters) != 2 {
		t.Fatalf("expected 2 items, got %d", len(adapters))
	}

	for i, adapter := range adapters {
		orig := objects[i].(*astral.Int8)

		if adapter.Type != "int8" {
			t.Fatalf("expected Type=int8, got %s", adapter.Type)
		}

		// Ensure adapter.Object contains JSON, not pointer reuse
		objBytes, _ := json.Marshal(adapter.Object)
		if len(objBytes) == 0 {
			t.Fatalf("expected Object field, got empty")
		}

		// decode into new primitive
		var decoded astral.Int8
		decodedBytes, _ := json.Marshal(adapter.Object)
		_ = json.Unmarshal(decodedBytes, &decoded)

		if decoded != *orig {
			t.Fatalf("expected %d, got %d", *orig, decoded)
		}

		if adapter.Object == orig {
			t.Fatalf("expected JSON-mapped value, got original pointer")
		}
	}

	var decodedSlice astral.Slice[astral.Object]
	if err := decodedSlice.UnmarshalJSON(data); err != nil {
		t.Fatalf("round-trip Unmarshal failed: %v", err)
	}

	if len(*decodedSlice.Elem) != 2 {
		t.Fatalf("round-trip len mismatch: expected 2, got %d", len(*decodedSlice.Elem))
	}

	if int8(*(*decodedSlice.Elem)[0].(*astral.Int8)) != 11 {
		t.Fatalf("round-trip mismatch: got %d at index 0", *(*decodedSlice.Elem)[0].(*astral.Int8))
	}

	if int8(*(*decodedSlice.Elem)[1].(*astral.Int8)) != 99 {
		t.Fatalf("round-trip mismatch: got %d at index 1", *(*decodedSlice.Elem)[1].(*astral.Int8))
	}
}

func TestSlice_UnmarshalJSON_Int8Only(t *testing.T) {
	// JSON matching what Slice.MarshalJSON emits for two Int8 values.
	jsonData := `[
		{"Type":"int8","Object":11},
		{"Type":"int8","Object":99}
	]`

	// backing slice for the result
	var out []astral.Object
	slice := astral.WrapSlice(&out)

	// decode
	if err := slice.UnmarshalJSON([]byte(jsonData)); err != nil {
		t.Fatalf("UnmarshalJSON failed: %v", err)
	}

	// basic length check
	if len(out) != 2 {
		t.Fatalf("expected 2 items, got %d", len(out))
	}

	// check element 0
	i0, ok := out[0].(*astral.Int8)
	if !ok {
		t.Fatalf("index 0: expected *astral.Int8, got %T", out[0])
	}
	if *i0 != astral.Int8(11) {
		t.Fatalf("index 0: expected 11, got %d", *i0)
	}

	// check element 1
	i1, ok := out[1].(*astral.Int8)
	if !ok {
		t.Fatalf("index 1: expected *astral.Int8, got %T", out[1])
	}
	if *i1 != astral.Int8(99) {
		t.Fatalf("index 1: expected 99, got %d", *i1)
	}
}

func TestSlice_UnmarshalJSON_TestA_TestB(t *testing.T) {
	astral.DefaultBlueprints.Add(&TestObjectA{})
	astral.DefaultBlueprints.Add(&TestObjectB{})

	// This JSON matches what MarshalJSON produces for TestObjectA/TestObjectB
	jsonData := `[
        {"Type":"test.object.a","Object":{"Value":"hello"}},
        {"Type":"test.object.b","Object":{"Number":"99"}}
    ]`

	var out []astral.Object
	slice := astral.WrapSlice(&out)

	if err := slice.UnmarshalJSON([]byte(jsonData)); err != nil {
		t.Fatalf("UnmarshalJSON failed: %v", err)
	}

	if len(out) != 2 {
		t.Fatalf("expected 2 items, got %d", len(out))
	}

	// ----- TestObjectA -----
	a, ok := out[0].(*TestObjectA)
	if !ok {
		t.Fatalf("index 0: expected *TestObjectA, got %T", out[0])
	}
	if a.Value != astral.String("hello") {
		t.Fatalf("index 0: expected Value='hello', got '%s'", a.Value)
	}

	// ----- TestObjectB -----
	b, ok := out[1].(*TestObjectB)
	if !ok {
		t.Fatalf("index 1: expected *TestObjectB, got %T", out[1])
	}
	if b.Number != astral.Int8(99) {
		t.Fatalf("index 1: expected Number=99, got %d", b.Number)
	}
}
