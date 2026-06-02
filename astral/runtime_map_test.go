package astral

import (
	"bytes"
	"encoding/hex"
	"strings"
	"testing"
)

func TestRuntimeMap_HeterogeneousStringKeyRoundTrip(t *testing.T) {
	src, err := newRuntimeMap("string16", "")
	if err != nil {
		t.Fatal(err)
	}
	if err := src.Set("a", NewUint32(1)); err != nil {
		t.Fatal(err)
	}
	if err := src.Set("z", NewString16("hi")); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	if _, err := src.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}

	dst, err := newRuntimeMap("string16", "")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := dst.ReadFrom(&buf); err != nil {
		t.Fatal(err)
	}

	if dst.Len() != 2 {
		t.Fatalf("len: want 2, got %d", dst.Len())
	}
	a, ok := dst.Get("a")
	if !ok {
		t.Fatal("missing key a")
	}
	if u, ok := a.(*Uint32); !ok || *u != 1 {
		t.Fatalf("a: want *Uint32(1), got %T %v", a, a)
	}
	z, ok := dst.Get("z")
	if !ok {
		t.Fatal("missing key z")
	}
	if s, ok := z.(*String16); !ok || *s != "hi" {
		t.Fatalf("z: want *String16(\"hi\"), got %T %v", z, z)
	}
}

func TestRuntimeMap_HomogeneousStringKeyRoundTrip(t *testing.T) {
	src, err := newRuntimeMap("string16", "uint32")
	if err != nil {
		t.Fatal(err)
	}
	for k, v := range map[string]uint32{"a": 1, "b": 2, "c": 3} {
		if err := src.Set(k, NewUint32(v)); err != nil {
			t.Fatal(err)
		}
	}

	var buf bytes.Buffer
	if _, err := src.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}

	dst, err := newRuntimeMap("string16", "uint32")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := dst.ReadFrom(&buf); err != nil {
		t.Fatal(err)
	}
	if dst.Len() != 3 {
		t.Fatalf("len: want 3, got %d", dst.Len())
	}
	for k, want := range map[string]uint32{"a": 1, "b": 2, "c": 3} {
		v, ok := dst.Get(k)
		if !ok {
			t.Fatalf("missing key %q", k)
		}
		u, ok := v.(*Uint32)
		if !ok || *u != Uint32(want) {
			t.Fatalf("[%q]: want %d, got %T %v", k, want, v, v)
		}
	}
}

func TestRuntimeMap_HeterogeneousUint16KeyRoundTrip(t *testing.T) {
	src, err := newRuntimeMap("uint16", "")
	if err != nil {
		t.Fatal(err)
	}
	if err := src.Set(uint64(1), NewUint32(42)); err != nil {
		t.Fatal(err)
	}
	if err := src.Set(uint64(256), NewString16("hi")); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	if _, err := src.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}

	dst, err := newRuntimeMap("uint16", "")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := dst.ReadFrom(&buf); err != nil {
		t.Fatal(err)
	}

	v, ok := dst.Get(uint64(1))
	if !ok {
		t.Fatal("missing key 1")
	}
	if u, ok := v.(*Uint32); !ok || *u != 42 {
		t.Fatalf("[1]: want *Uint32(42), got %T %v", v, v)
	}
	v, ok = dst.Get(uint64(256))
	if !ok {
		t.Fatal("missing key 256")
	}
	if s, ok := v.(*String16); !ok || *s != "hi" {
		t.Fatalf("[256]: want *String16(\"hi\"), got %T %v", v, v)
	}
}

func TestRuntimeMap_HomogeneousUint16KeyRoundTrip(t *testing.T) {
	src, err := newRuntimeMap("uint16", "uint32")
	if err != nil {
		t.Fatal(err)
	}
	for k, v := range map[uint64]uint32{1: 10, 256: 20} {
		if err := src.Set(k, NewUint32(v)); err != nil {
			t.Fatal(err)
		}
	}

	var buf bytes.Buffer
	if _, err := src.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}

	dst, err := newRuntimeMap("uint16", "uint32")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := dst.ReadFrom(&buf); err != nil {
		t.Fatal(err)
	}
	for k, want := range map[uint64]uint32{1: 10, 256: 20} {
		v, ok := dst.Get(k)
		if !ok {
			t.Fatalf("missing key %d", k)
		}
		u, ok := v.(*Uint32)
		if !ok || *u != Uint32(want) {
			t.Fatalf("[%d]: want %d, got %T %v", k, want, v, v)
		}
	}
}

func TestRuntimeMap_StringKeyDeterministic(t *testing.T) {
	build := func() *runtimeMap {
		m, _ := newRuntimeMap("string16", "uint32")
		_ = m.Set("c", NewUint32(3))
		_ = m.Set("a", NewUint32(1))
		_ = m.Set("b", NewUint32(2))
		return m
	}

	var b1, b2 bytes.Buffer
	if _, err := build().WriteTo(&b1); err != nil {
		t.Fatal(err)
	}
	if _, err := build().WriteTo(&b2); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(b1.Bytes(), b2.Bytes()) {
		t.Fatal("encoding not deterministic")
	}
	// skip 4-byte count + 2-byte String16 length = 6 bytes, then first key byte must be 'a'
	if b1.Bytes()[6] != 'a' {
		t.Fatalf("expected first key 'a', got %q", b1.Bytes()[6])
	}
}

func TestRuntimeMap_Uint16KeyDeterministic(t *testing.T) {
	build := func() *runtimeMap {
		m, _ := newRuntimeMap("uint16", "uint32")
		_ = m.Set(uint64(0x0300), NewUint32(3))
		_ = m.Set(uint64(0x0100), NewUint32(1))
		_ = m.Set(uint64(0x0200), NewUint32(2))
		return m
	}

	var b1, b2 bytes.Buffer
	if _, err := build().WriteTo(&b1); err != nil {
		t.Fatal(err)
	}
	if _, err := build().WriteTo(&b2); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(b1.Bytes(), b2.Bytes()) {
		t.Fatal("encoding not deterministic")
	}
	first := uint16(b1.Bytes()[4])<<8 | uint16(b1.Bytes()[5])
	if first != 0x0100 {
		t.Fatalf("expected first key 0x0100, got %x", first)
	}
}

func TestRuntimeMap_KeyWidthOverflow(t *testing.T) {
	m, err := newRuntimeMap("uint16", "uint32")
	if err != nil {
		t.Fatal(err)
	}
	if err := m.Set(uint64(1<<16), NewUint32(0)); err == nil {
		t.Fatal("expected width overflow error")
	}
}

func TestRuntimeMap_KeyKindRejection(t *testing.T) {
	sm, _ := newRuntimeMap("string16", "uint32")
	if err := sm.Set(uint64(1), NewUint32(0)); err == nil {
		t.Fatal("expected string-key map to reject uint64 key")
	}
	im, _ := newRuntimeMap("uint16", "uint32")
	if err := im.Set("a", NewUint32(0)); err == nil {
		t.Fatal("expected uint-key map to reject string key")
	}
}

func TestRuntimeMap_ValueTypeRejection(t *testing.T) {
	m, err := newRuntimeMap("string16", "uint32")
	if err != nil {
		t.Fatal(err)
	}
	if err := m.Set("a", NewString16("nope")); err == nil {
		t.Fatal("expected value-type mismatch error")
	}
}

func TestRuntimeMap_UnsupportedKeyType(t *testing.T) {
	if _, err := newRuntimeMap("int32", ""); err == nil {
		t.Fatal("expected unsupported-key-type error for int32")
	}
	if _, err := newRuntimeMap("uint", ""); err == nil {
		t.Fatal("expected unsupported-key-type error for plain uint")
	}
}

func TestRuntimeMap_EmptyRoundTrip(t *testing.T) {
	src, err := newRuntimeMap("string16", "uint32")
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if _, err := src.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}

	dst, err := newRuntimeMap("string16", "uint32")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := dst.ReadFrom(&buf); err != nil {
		t.Fatal(err)
	}
	if dst.Len() != 0 {
		t.Fatalf("len: want 0, got %d", dst.Len())
	}
}

func TestRuntimeMap_JSONRoundTrip_StringKey(t *testing.T) {
	src, err := newRuntimeMap("string16", "uint32")
	if err != nil {
		t.Fatal(err)
	}
	_ = src.Set("a", NewUint32(1))
	_ = src.Set("b", NewUint32(2))

	data, err := src.MarshalJSON()
	if err != nil {
		t.Fatal(err)
	}

	dst, err := newRuntimeMap("string16", "uint32")
	if err != nil {
		t.Fatal(err)
	}
	if err := dst.UnmarshalJSON(data); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if dst.Len() != 2 {
		t.Fatalf("len after json: want 2, got %d", dst.Len())
	}
}

func TestRuntimeMap_JSONRoundTrip_UintKey(t *testing.T) {
	src, err := newRuntimeMap("uint16", "uint32")
	if err != nil {
		t.Fatal(err)
	}
	_ = src.Set(uint64(1), NewUint32(10))
	_ = src.Set(uint64(256), NewUint32(20))

	data, err := src.MarshalJSON()
	if err != nil {
		t.Fatal(err)
	}
	// raw entries dict; no width envelope (D3)
	if strings.Contains(string(data), `"width"`) {
		t.Fatalf("JSON should not contain a width envelope: %s", data)
	}

	dst, err := newRuntimeMap("uint16", "uint32")
	if err != nil {
		t.Fatal(err)
	}
	if err := dst.UnmarshalJSON(data); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if dst.Len() != 2 {
		t.Fatalf("len after json: want 2, got %d", dst.Len())
	}
}

// TestRuntimeMap_CrossCodecParity_StringHomogeneous pins the wire-format equivalence between a
// native Go map field walked by Objectify and a runtimeMap built from a MapSpec with the same
// element type. Any drift in either codec breaks this test.
func TestRuntimeMap_CrossCodecParity_StringHomogeneous(t *testing.T) {
	native := map[string]*Uint32{"a": NewUint32(1), "bb": NewUint32(2)}

	var nativeBuf bytes.Buffer
	if _, err := Objectify(&native).WriteTo(&nativeBuf); err != nil {
		t.Fatal(err)
	}

	rm, err := newRuntimeMap("string16", "uint32")
	if err != nil {
		t.Fatal(err)
	}
	for k, v := range native {
		if err := rm.Set(k, v); err != nil {
			t.Fatal(err)
		}
	}

	var rmBuf bytes.Buffer
	if _, err := rm.WriteTo(&rmBuf); err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(nativeBuf.Bytes(), rmBuf.Bytes()) {
		t.Fatalf("parity mismatch:\n native: %x\n  rmap: %x", nativeBuf.Bytes(), rmBuf.Bytes())
	}
}

// TestRuntimeMap_LockedBytes_StringHomogeneous pins the concrete wire bytes for a
// MapSpec{KeyType:"string16", ValueType:"uint32"} map containing {"a"→42, "bb"→99}. Any
// future change to the homogeneous wire shape (nilflag presence, sort order, key length
// prefixes) breaks this test and must be a conscious decision.
func TestRuntimeMap_LockedBytes_StringHomogeneous(t *testing.T) {
	rm, err := newRuntimeMap("string16", "uint32")
	if err != nil {
		t.Fatal(err)
	}
	_ = rm.Set("a", NewUint32(42))
	_ = rm.Set("bb", NewUint32(99))

	var buf bytes.Buffer
	if _, err := rm.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}

	want, _ := hex.DecodeString(
		"00000002" + // count = 2
			"0001" + "61" + // key "a"
			"01" + // nilflag
			"0000002a" + // uint32 = 42
			"0002" + "6262" + // key "bb"
			"01" + // nilflag
			"00000063", // uint32 = 99
	)
	if !bytes.Equal(buf.Bytes(), want) {
		t.Fatalf("locked-bytes mismatch:\n  got: %x\n want: %x", buf.Bytes(), want)
	}
}
