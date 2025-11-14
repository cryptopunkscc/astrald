package astral_test

import (
	"encoding/json"
	"io"
	"testing"

	"github.com/cryptopunkscc/astrald/astral"
)

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

func TestSlice_MarshalJSON_PrimitivesOnly(t *testing.T) {
	objects := []astral.Object{
		P(astral.String("hello")),
		P(astral.String("world")),
	}
	slice := astral.WrapSlice(&objects)

	data, err := slice.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON returned error: %v", err)
	}

	var adapters []astral.JSONEncodeAdapter
	if err := json.Unmarshal(data, &adapters); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if len(adapters) != 2 {
		t.Fatalf("expected 2 items, got %d", len(adapters))
	}

	for i, adapter := range adapters {
		if adapter.Type != "string" {
			t.Fatalf("expected type string, got %s", adapter.Type)
		}

		expectedStr, ok := objects[i].(*astral.String)
		if !ok {
			t.Fatalf("expected string, got %T", adapter.Object)
		}

		if adapter.Object == expectedStr {
			t.Fatalf("expected object %v, got %v", objects[i], adapter.Object)
		}
	}
}

func TestSlice_MarshalJSON_MixedTypes(t *testing.T) {
	objects := []astral.Object{
		P(astral.String("hello")),
		P(astral.Int8(42)),
	}

	slice := astral.WrapSlice(&objects)

	data, err := slice.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON returned error: %v", err)
	}

	var adapters []astral.JSONEncodeAdapter
	if err := json.Unmarshal(data, &adapters); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if len(adapters) != 2 {
		t.Fatalf("expected 2 items, got %d", len(adapters))
	}

	for i, adapter := range adapters {
		expectedType := ""
		var expectedPtr astral.Object

		switch obj := objects[i].(type) {
		case *astral.String:
			expectedType = "string"
			expectedPtr = obj

		case *astral.Int8:
			expectedType = "int8"
			expectedPtr = obj

		default:
			t.Fatalf("unexpected object type in test at index %d: %T", i, objects[i])
		}

		// type check
		if adapter.Type != expectedType {
			t.Fatalf("expected type %s, got %s", expectedType, adapter.Type)
		}

		// never return original pointer in JSON
		if adapter.Object == expectedPtr {
			t.Fatalf("expected JSON-mapped object, got original pointer at index %d", i)
		}
	}
}

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
			if adapter.Object == orig {
				t.Fatalf("expected JSON object, got original pointer")
			}

		case *TestObjectB:
			if adapter.Type != "test.object.b" {
				t.Fatalf("expected type test.object.b, got %s", adapter.Type)
			}
			if adapter.Object == orig {
				t.Fatalf("expected JSON object, got original pointer")
			}

		default:
			t.Fatalf("unexpected object type at index %d: %T", i, objects[i])
		}
	}
}

func TestSlice_UnmarshalJSON_PrimitivesOnly(t *testing.T) {
	// Prepare JSON that matches output of MarshalJSON for two strings
	jsonData := `[
		{"type":"string","object":"hello"},
		{"type":"string","object":"world"}
	]`

	// Prepare target slice
	var objects []astral.Object
	slice := astral.WrapSlice(&objects)

	// Decode
	err := slice.UnmarshalJSON([]byte(jsonData))
	if err != nil {
		t.Fatalf("UnmarshalJSON returned error: %v", err)
	}

	// Validate slice length
	if len(objects) != 2 {
		t.Fatalf("expected 2 elements, got %d", len(objects))
	}

	// Validate elements
	tests := []string{"hello", "world"}

	for i, expected := range tests {
		strObj, ok := objects[i].(*astral.String)
		if !ok {
			t.Fatalf("element %d: expected *astral.String, got %T", i, objects[i])
		}

		if string(*strObj) != expected {
			t.Fatalf("element %d: expected %q, got %q", i, expected, string(*strObj))
		}
	}
}
func TestSlice_UnmarshalJSON_MixedTypes(t *testing.T) {
	original := []astral.Object{
		P(astral.String("hello")),
		P(astral.Int8(42)),
	}

	src := astral.WrapSlice(&original)
	jsonData, err := src.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON error: %v", err)
	}

	var decoded []astral.Object
	dst := astral.WrapSlice(&decoded)

	err = dst.UnmarshalJSON(jsonData)
	if err != nil {
		t.Fatalf("UnmarshalJSON returned error: %v", err)
	}

	if len(decoded) != 2 {
		t.Fatalf("expected 2 items, got %d", len(decoded))
	}

	str, ok := decoded[0].(*astral.String)
	if !ok {
		t.Fatalf("index 0: expected *astral.String, got %T", decoded[0])
	}
	if string(*str) != "hello" {
		t.Fatalf("expected \"hello\", got %q", string(*str))
	}

	i8, ok := decoded[1].(*astral.Int8)
	if !ok {
		t.Fatalf("index 1: expected *astral.Int8, got %T", decoded[1])
	}
	if int8(*i8) != 42 {
		t.Fatalf("expected 42, got %d", *i8)
	}
}

func TestSlice_UnmarshalJSON_CustomMixed(t *testing.T) {
	astral.DefaultBlueprints.Add(&TestObjectA{})
	astral.DefaultBlueprints.Add(&TestObjectB{})

	original := []astral.Object{
		&TestObjectA{Value: "hello"},
		&TestObjectB{Number: 99},
	}

	src := astral.WrapSlice(&original)
	jsonData, err := src.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON failed: %v", err)
	}

	var decoded []astral.Object
	dst := astral.WrapSlice(&decoded)

	if err := dst.UnmarshalJSON(jsonData); err != nil {
		t.Fatalf("UnmarshalJSON failed: %v", err)
	}

	if len(decoded) != len(original) {
		t.Fatalf("expected %d items, got %d", len(original), len(decoded))
	}

	for i := range decoded {
		switch orig := decoded[i].(type) {

		case *TestObjectA:
			if orig.Value != "hello" {
				t.Fatalf("index %d: expected Value=hello, got %s", i, orig.Value)
			}

		case *TestObjectB:
			if orig.Number != 99 {
				t.Fatalf("index %d: expected Number=99, got %d", i, orig.Number)
			}

		default:
			t.Fatalf("unexpected type at index %d: %T", i, decoded[i])
		}
	}
}
