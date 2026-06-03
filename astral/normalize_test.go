package astral

import (
	"strings"
	"testing"
)

// Regression tests for audit #13: narrowString/narrowBytes used to accept any input,
// deferring the length check to WriteTo (which was broken). The fix rejects at the
// API boundary so ro.Set returns immediately for oversized values.

func TestNarrowString_RejectsOversized(t *testing.T) {
	cases := []struct {
		name string
		len  int
	}{
		{"string8", 256},
		{"string16", 65536},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := narrowString(c.name, strings.Repeat("a", c.len))
			if err == nil {
				t.Fatalf("%s: want error for length %d, got nil", c.name, c.len)
			}
		})
	}
}

func TestNarrowString_AcceptsExactlyMax(t *testing.T) {
	cases := []struct {
		name string
		len  int
	}{
		{"string8", 255},
		{"string16", 65535},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := narrowString(c.name, strings.Repeat("a", c.len))
			if err != nil {
				t.Fatalf("%s: unexpected error at max length %d: %v", c.name, c.len, err)
			}
		})
	}
}

func TestNarrowBytes_RejectsOversized(t *testing.T) {
	cases := []struct {
		name string
		len  int
	}{
		{"bytes8", 256},
		{"bytes16", 65536},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := narrowBytes(c.name, make([]byte, c.len))
			if err == nil {
				t.Fatalf("%s: want error for length %d, got nil", c.name, c.len)
			}
		})
	}
}

func TestNarrowBytes_AcceptsExactlyMax(t *testing.T) {
	cases := []struct {
		name string
		len  int
	}{
		{"bytes8", 255},
		{"bytes16", 65535},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := narrowBytes(c.name, make([]byte, c.len))
			if err != nil {
				t.Fatalf("%s: unexpected error at max length %d: %v", c.name, c.len, err)
			}
		})
	}
}

// TestRuntimeObject_Set_RejectsOversizedString exercises the end-to-end path:
// ro.Set → normalize → narrowString. Before the fix this accepted oversized values
// which would later corrupt the wire on WriteTo.
func TestRuntimeObject_Set_RejectsOversizedString(t *testing.T) {
	bp := NewBlueprint("test.oversize.string",
		Field{Name: "s", Spec: &PrimitiveSpec{PrimitiveType: "string8"}},
	)
	ro, err := NewRuntimeObject(bp)
	if err != nil {
		t.Fatal(err)
	}
	err = ro.Set("s", strings.Repeat("a", 256))
	if err == nil {
		t.Fatal("want error for 256-byte string into string8 field, got nil")
	}
}

func TestNormalize_ArraySpecLengthMismatch_Message(t *testing.T) {
	spec := &ArraySpec{Type: "uint32", Length: 5}
	ra, err := NewRuntimeArray("uint32", 3)
	if err != nil {
		t.Fatal(err)
	}
	_, err = normalize(spec, ra)
	if err == nil {
		t.Fatal("want error for length mismatch, got nil")
	}
	if !strings.Contains(err.Error(), "ArraySpec") {
		t.Fatalf("want error to name ArraySpec, got %v", err)
	}
	for _, want := range []string{"5", "3"} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("want error to include %q (expected vs actual length), got %v", want, err)
		}
	}
}

// TestNormalize_RejectsNilForNonNullable — §13.4. RefSpec and ObjectSpec are not
// nullable; passing nil through normalize must surface a clear "does not accept nil"
// error so callers can distinguish "no value" from "value mismatch".
func TestNormalize_RejectsNilForNonNullable(t *testing.T) {
	t.Run("RefSpec", func(t *testing.T) {
		spec := &RefSpec{Type: "astral.blueprint"}
		_, err := normalize(spec, nil)
		if err == nil {
			t.Fatal("want error for nil into RefSpec, got nil")
		}
		if !strings.Contains(err.Error(), "does not accept nil") {
			t.Fatalf("want \"does not accept nil\" in error, got %v", err)
		}
	})
	t.Run("ObjectSpec", func(t *testing.T) {
		spec := &ObjectSpec{}
		_, err := normalize(spec, nil)
		if err == nil {
			t.Fatal("want error for nil into ObjectSpec, got nil")
		}
		if !strings.Contains(err.Error(), "does not accept nil") {
			t.Fatalf("want \"does not accept nil\" in error, got %v", err)
		}
	})
}

func TestAdapt_TypedNilPointer(t *testing.T) {
	cases := []struct {
		name string
		v    any
	}{
		{"int", (*int)(nil)},
		{"uint32", (*uint32)(nil)},
		{"string", (*string)(nil)},
		{"object", (Object)(nil)},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("Adapt panicked on typed-nil %s: %v", c.name, r)
				}
			}()
			got := Adapt(c.v)
			if got == nil {
				t.Fatalf("Adapt on typed-nil %s returned nil; want &Nil{}", c.name)
			}
			if _, ok := got.(*Nil); !ok {
				t.Fatalf("Adapt on typed-nil %s: want *Nil, got %T", c.name, got)
			}
		})
	}
}
