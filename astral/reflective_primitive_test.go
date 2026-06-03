package astral

import (
	"bytes"
	"io"
	"reflect"
	"testing"
)

// TestReflectivePrimitive_ReadFromPropagatesShortRead pins the contract for every
// reflective primitive *Value.ReadFrom: a short or failed read must return (0, err),
// never (width, nil). Previously each decoder discarded the binary.Read error and
// reported success with a zero-valued field, silently corrupting struct/slice/map
// decode for every Objectify-built Object on a truncated stream.
func TestReflectivePrimitive_ReadFromPropagatesShortRead(t *testing.T) {
	cases := []struct {
		name string
		make func() Object
	}{
		{"uint8", func() Object { var v uint8; return uint8Value{reflect.ValueOf(&v).Elem()} }},
		{"uint16", func() Object { var v uint16; return uint16Value{reflect.ValueOf(&v).Elem()} }},
		{"uint32", func() Object { var v uint32; return uint32Value{reflect.ValueOf(&v).Elem()} }},
		{"uint64", func() Object { var v uint64; return uint64Value{reflect.ValueOf(&v).Elem()} }},
		{"int8", func() Object { var v int8; return int8Value{reflect.ValueOf(&v).Elem()} }},
		{"int16", func() Object { var v int16; return int16Value{reflect.ValueOf(&v).Elem()} }},
		{"int32", func() Object { var v int32; return int32Value{reflect.ValueOf(&v).Elem()} }},
		{"int64", func() Object { var v int64; return int64Value{reflect.ValueOf(&v).Elem()} }},
		{"float32", func() Object { var v float32; return float32Value{reflect.ValueOf(&v).Elem()} }},
		{"float64", func() Object { var v float64; return float64Value{reflect.ValueOf(&v).Elem()} }},
	}

	for _, c := range cases {
		t.Run(c.name+"/empty", func(t *testing.T) {
			obj := c.make()
			n, err := obj.ReadFrom(bytes.NewReader(nil))
			if err == nil {
				t.Fatalf("want error on empty reader, got n=%d err=nil", n)
			}
			if n != 0 {
				t.Fatalf("want n=0 on failed read, got n=%d", n)
			}
		})

		t.Run(c.name+"/short", func(t *testing.T) {
			// One byte — too short for every width except uint8/int8, which the empty
			// case already covers. Skip those here; for ≥2-byte widths this surfaces
			// io.ErrUnexpectedEOF.
			if c.name == "uint8" || c.name == "int8" {
				return
			}
			obj := c.make()
			n, err := obj.ReadFrom(bytes.NewReader([]byte{0xFF}))
			if err == nil {
				t.Fatalf("want error on 1-byte reader for %s, got n=%d err=nil", c.name, n)
			}
			if err != io.ErrUnexpectedEOF && err != io.EOF {
				t.Fatalf("want io.ErrUnexpectedEOF or io.EOF for %s, got %v", c.name, err)
			}
			if n != 0 {
				t.Fatalf("want n=0 on failed read for %s, got n=%d", c.name, n)
			}
		})
	}
}
