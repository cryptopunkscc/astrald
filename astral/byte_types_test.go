package astral

import (
	"bytes"
	"testing"
)

// Regression tests for audit #13: same lossy-cast bug in Bytes8/16 as in String8/16.

func TestBytes8_WriteTo_RejectsOversized(t *testing.T) {
	b := Bytes8(make([]byte, 256))
	var buf bytes.Buffer
	_, err := b.WriteTo(&buf)
	if err == nil {
		t.Fatalf("want error for 256-byte Bytes8, got nil; wrote %d bytes", buf.Len())
	}
	if buf.Len() != 0 {
		t.Fatalf("expected no bytes written on error, got %d", buf.Len())
	}
}

func TestBytes8_WriteTo_AcceptsExactlyMax(t *testing.T) {
	b := Bytes8(make([]byte, 255))
	var buf bytes.Buffer
	_, err := b.WriteTo(&buf)
	if err != nil {
		t.Fatal(err)
	}
	if buf.Len() != 256 {
		t.Fatalf("want 1+255=256 bytes written, got %d", buf.Len())
	}
}

func TestBytes16_WriteTo_RejectsOversized(t *testing.T) {
	b := Bytes16(make([]byte, 65536))
	var buf bytes.Buffer
	_, err := b.WriteTo(&buf)
	if err == nil {
		t.Fatalf("want error for 65536-byte Bytes16, got nil")
	}
	if buf.Len() != 0 {
		t.Fatalf("expected no bytes written on error, got %d", buf.Len())
	}
}

func TestBytes16_WriteTo_AcceptsExactlyMax(t *testing.T) {
	b := Bytes16(make([]byte, 65535))
	var buf bytes.Buffer
	_, err := b.WriteTo(&buf)
	if err != nil {
		t.Fatal(err)
	}
	if buf.Len() != 2+65535 {
		t.Fatalf("want 2+65535 bytes written, got %d", buf.Len())
	}
}
