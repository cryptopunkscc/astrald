package astral

import (
	"bytes"
	"testing"
)

func TestRuntimeSlice_HeterogeneousRoundTrip(t *testing.T) {
	src, err := newRuntimeSlice("")
	if err != nil {
		t.Fatal(err)
	}
	if err := src.Append(NewUint32(1)); err != nil {
		t.Fatal(err)
	}
	if err := src.Append(NewString16("hi")); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	if _, err := src.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}

	dst, err := newRuntimeSlice("")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := dst.ReadFrom(&buf); err != nil {
		t.Fatal(err)
	}
	if dst.Len() != 2 {
		t.Fatalf("len: want 2, got %d", dst.Len())
	}
	if u, ok := dst.At(0).(*Uint32); !ok || *u != 1 {
		t.Fatalf("[0]: want *Uint32(1), got %#v", dst.At(0))
	}
	if s, ok := dst.At(1).(*String16); !ok || *s != "hi" {
		t.Fatalf("[1]: want *String16(\"hi\"), got %#v", dst.At(1))
	}
}

func TestRuntimeSlice_HomogeneousRoundTrip(t *testing.T) {
	src, err := newRuntimeSlice("uint32")
	if err != nil {
		t.Fatal(err)
	}
	for _, v := range []uint32{1, 2, 3} {
		if err := src.Append(NewUint32(v)); err != nil {
			t.Fatal(err)
		}
	}

	var buf bytes.Buffer
	if _, err := src.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}

	dst, err := newRuntimeSlice("uint32")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := dst.ReadFrom(&buf); err != nil {
		t.Fatal(err)
	}
	if dst.Len() != 3 {
		t.Fatalf("len: want 3, got %d", dst.Len())
	}
	for i, want := range []uint32{1, 2, 3} {
		u, ok := dst.At(i).(*Uint32)
		if !ok || *u != Uint32(want) {
			t.Fatalf("[%d]: want %d, got %#v", i, want, dst.At(i))
		}
	}
}

func TestRuntimeSlice_AppendTypeRejection(t *testing.T) {
	s, err := newRuntimeSlice("uint32")
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Append(NewString16("x")); err == nil {
		t.Fatal("expected type mismatch error, got nil")
	}
}

func TestRuntimeSlice_EmptyRoundTrip(t *testing.T) {
	src, err := newRuntimeSlice("uint32")
	if err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	if _, err := src.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}

	dst, err := newRuntimeSlice("uint32")
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

func TestRuntimeSlice_CrossCodecParity_Homogeneous(t *testing.T) {
	homo := []*Uint32{NewUint32(1), NewUint32(2)}

	var nativeBuf bytes.Buffer
	if _, err := Objectify(&homo).WriteTo(&nativeBuf); err != nil {
		t.Fatal(err)
	}

	rs, err := newRuntimeSlice("uint32")
	if err != nil {
		t.Fatal(err)
	}
	for _, v := range homo {
		if err := rs.Append(v); err != nil {
			t.Fatal(err)
		}
	}

	var rsBuf bytes.Buffer
	if _, err := rs.WriteTo(&rsBuf); err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(nativeBuf.Bytes(), rsBuf.Bytes()) {
		t.Fatalf("homogeneous parity mismatch:\n native: %x\n  slice: %x", nativeBuf.Bytes(), rsBuf.Bytes())
	}
}

func TestRuntimeSlice_CrossCodecParity_Heterogeneous(t *testing.T) {
	hetero := []Object{NewUint32(1), NewString16("hi")}

	var nativeBuf bytes.Buffer
	if _, err := Objectify(&hetero).WriteTo(&nativeBuf); err != nil {
		t.Fatal(err)
	}

	rs, err := newRuntimeSlice("")
	if err != nil {
		t.Fatal(err)
	}
	for _, v := range hetero {
		if err := rs.Append(v); err != nil {
			t.Fatal(err)
		}
	}

	var rsBuf bytes.Buffer
	if _, err := rs.WriteTo(&rsBuf); err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(nativeBuf.Bytes(), rsBuf.Bytes()) {
		t.Fatalf("heterogeneous parity mismatch:\n native: %x\n  slice: %x", nativeBuf.Bytes(), rsBuf.Bytes())
	}
}
