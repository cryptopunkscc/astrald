package astral

import (
	"bytes"
	"testing"
)

type testMap struct {
	SomeMap   map[Uint8]String8
	NativeMap map[uint8]string
}

func TestMapWithValues(t *testing.T) {
	var err error
	var src, dst testMap

	src.SomeMap = make(map[Uint8]String8)
	src.SomeMap[1] = "hello world"
	src.NativeMap = make(map[uint8]string)
	src.NativeMap[1] = "hello world"

	srcObject := Objectify(&src)

	// test binary marshaling
	var buf = &bytes.Buffer{}
	_, err = srcObject.WriteTo(buf)
	if err != nil {
		t.Fatal(err)
	}

	dst.SomeMap = nil
	dstObject := Objectify(&dst)
	_, err = dstObject.ReadFrom(buf)
	if err != nil {
		t.Fatal(err)
	}

	switch {
	case len(dst.SomeMap) != len(src.SomeMap):
		t.Fatal("map lengths not equal")
	case dst.SomeMap[1] != src.SomeMap[1]:
		t.Fatal("map values not equal")
	case len(dst.NativeMap) != len(src.NativeMap):
		t.Fatal("native map lengths not equal")
	case dst.NativeMap[1] != src.NativeMap[1]:
		t.Fatal("native map values not equal")
	}

	// test json marshaling
	jdata, err := srcObject.MarshalJSON()
	if err != nil {
		t.Fatal(err)
	}

	dst = testMap{}
	dstObject = Objectify(&dst)

	err = dstObject.UnmarshalJSON(jdata)
	if err != nil {
		t.Fatal(err)
	}

	switch {
	case len(dst.SomeMap) != len(src.SomeMap):
		t.Fatal("map lengths not equal")
	case dst.SomeMap[1] != src.SomeMap[1]:
		t.Fatal("map values not equal")
	case len(dst.NativeMap) != len(src.NativeMap):
		t.Fatal("native map lengths not equal")
	case dst.NativeMap[1] != src.NativeMap[1]:
		t.Fatal("native map values not equal")
	}
}

func TestMapEmpty(t *testing.T) {
	var err error
	var src, dst testMap

	srcObject := Objectify(&src)

	// test binary marshaling
	var buf = &bytes.Buffer{}
	_, err = srcObject.WriteTo(buf)
	if err != nil {
		t.Fatal(err)
	}

	dst.SomeMap = make(map[Uint8]String8)
	dst.SomeMap[1] = "hello world"

	dstObject := Objectify(&dst)
	_, err = dstObject.ReadFrom(buf)
	if err != nil {
		t.Fatal(err)
	}

	if len(dst.SomeMap) != 0 {
		t.Fatal("invalid map length")
	}

	// test json marshaling
	dst.SomeMap = make(map[Uint8]String8)
	dst.SomeMap[1] = "hello world"

	jdata, err := srcObject.MarshalJSON()
	if err != nil {
		t.Fatal(err)
	}

	dstObject = Objectify(&dst)
	err = dstObject.UnmarshalJSON(jdata)
	if err != nil {
		t.Fatal(err)
	}

	if len(dst.SomeMap) != 0 {
		t.Fatal("invalid map length")
	}
}

// TestMap_Empty_RoundTrip — §5.1. Empty maps must round-trip across string-keyed and
// int-keyed variants — the count=0 short-circuit on the decode side.
func TestMap_Empty_RoundTrip(t *testing.T) {
	t.Run("string_uint32", func(t *testing.T) {
		var src, dst struct {
			M map[string]uint32
		}
		dst.M = map[string]uint32{"keep": 1}

		var buf bytes.Buffer
		if _, err := Objectify(&src).WriteTo(&buf); err != nil {
			t.Fatal(err)
		}
		if _, err := Objectify(&dst).ReadFrom(&buf); err != nil {
			t.Fatal(err)
		}
		if len(dst.M) != 0 {
			t.Fatalf("want empty map, got %v", dst.M)
		}
	})
	t.Run("uint32_uint32", func(t *testing.T) {
		var src, dst struct {
			M map[uint32]uint32
		}
		dst.M = map[uint32]uint32{1: 2}

		var buf bytes.Buffer
		if _, err := Objectify(&src).WriteTo(&buf); err != nil {
			t.Fatal(err)
		}
		if _, err := Objectify(&dst).ReadFrom(&buf); err != nil {
			t.Fatal(err)
		}
		if len(dst.M) != 0 {
			t.Fatalf("want empty map, got %v", dst.M)
		}
	})
}

// TestMap_Encoding_ByteDeterministic — §5.4. Pairs are sorted by encoded key bytes, so
// repeated encoding of the same map must produce byte-identical output regardless of Go
// map iteration order.
func TestMap_Encoding_ByteDeterministic(t *testing.T) {
	src := struct {
		M map[string]uint32
	}{
		M: map[string]uint32{
			"alpha":   1,
			"bravo":   2,
			"charlie": 3,
			"delta":   4,
			"echo":    5,
		},
	}

	var first bytes.Buffer
	if _, err := Objectify(&src).WriteTo(&first); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 20; i++ {
		var buf bytes.Buffer
		if _, err := Objectify(&src).WriteTo(&buf); err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(first.Bytes(), buf.Bytes()) {
			t.Fatalf("iteration %d: non-deterministic encoding\nfirst: %x\n  now: %x",
				i, first.Bytes(), buf.Bytes())
		}
	}
}

// TestMap_PointerReceiverValue — §5.5. Map value type whose ObjectType is defined on the
// pointer receiver works because mapValue.WriteTo addresses the value via addressableMapValue.
// Without that workaround the reflect.Value from MapIndex is non-addressable and
// pointer-receiver methods would not be discoverable through objectify.
func TestMap_PointerReceiverValue(t *testing.T) {
	src := struct {
		M map[string]*String8
	}{
		M: map[string]*String8{
			"a": NewString8("alpha"),
			"b": NewString8("bravo"),
		},
	}
	var dst struct {
		M map[string]*String8
	}
	var buf bytes.Buffer
	if _, err := Objectify(&src).WriteTo(&buf); err != nil {
		t.Fatal(err)
	}
	if _, err := Objectify(&dst).ReadFrom(&buf); err != nil {
		t.Fatal(err)
	}
	if got, want := len(dst.M), len(src.M); got != want {
		t.Fatalf("len: want %d, got %d", want, got)
	}
	if got := dst.M["a"]; got == nil || string(*got) != "alpha" {
		t.Fatalf("dst.M[a]: %v", got)
	}
	if got := dst.M["b"]; got == nil || string(*got) != "bravo" {
		t.Fatalf("dst.M[b]: %v", got)
	}
}

// TestMap_NilInterfaceValue — §5.8. A map keyed by string with Object-valued (nil) entries
// must round-trip; the interfaceValue codec emits a zero-length type tag for nil.
func TestMap_NilInterfaceValue(t *testing.T) {
	src := struct {
		M map[string]Object
	}{
		M: map[string]Object{
			"absent":  nil,
			"present": NewString16("hi"),
		},
	}
	var dst struct {
		M map[string]Object
	}
	var buf bytes.Buffer
	if _, err := Objectify(&src).WriteTo(&buf); err != nil {
		t.Fatal(err)
	}
	if _, err := Objectify(&dst).ReadFrom(&buf); err != nil {
		t.Fatal(err)
	}
	if got, want := len(dst.M), 2; got != want {
		t.Fatalf("len: want %d, got %d", want, got)
	}
	if v := dst.M["absent"]; v != nil {
		t.Fatalf("absent: want nil, got %#v", v)
	}
	if s, ok := dst.M["present"].(*String16); !ok || *s != "hi" {
		t.Fatalf("present: want String16 hi, got %#v", dst.M["present"])
	}
}

func TestMap_UnmarshalJSON_NonNumericKey(t *testing.T) {
	var dst struct {
		M map[uint32]uint32
	}
	dstObj := Objectify(&dst)

	err := dstObj.UnmarshalJSON([]byte(`{"M":{"notanumber":1}}`))
	if err == nil {
		t.Fatal("want error for non-numeric key in integer-keyed map, got nil")
	}
}
