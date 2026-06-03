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
