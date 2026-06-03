package astral

import (
	"bytes"
	"strings"
	"testing"
)

// Regression tests for audit #13: the pre-fix code used `Uint8(len(s))` followed by an
// unreachable `l > (1<<8)-1` guard, so oversized strings wrote a truncated length followed
// by the full payload — desyncing the stream. Same shape in String16/32.

func TestString8_WriteTo_RejectsOversized(t *testing.T) {
	s := String8(strings.Repeat("a", 256))
	var buf bytes.Buffer
	_, err := s.WriteTo(&buf)
	if err == nil {
		t.Fatalf("want error for 256-byte String8, got nil; wrote %d bytes", buf.Len())
	}
	if buf.Len() != 0 {
		t.Fatalf("expected no bytes written on error, got %d", buf.Len())
	}
}

func TestString8_WriteTo_AcceptsExactlyMax(t *testing.T) {
	s := String8(strings.Repeat("a", 255))
	var buf bytes.Buffer
	_, err := s.WriteTo(&buf)
	if err != nil {
		t.Fatal(err)
	}
	if buf.Len() != 256 {
		t.Fatalf("want 1+255=256 bytes written, got %d", buf.Len())
	}
}

func TestString16_WriteTo_RejectsOversized(t *testing.T) {
	s := String16(strings.Repeat("a", 65536))
	var buf bytes.Buffer
	_, err := s.WriteTo(&buf)
	if err == nil {
		t.Fatalf("want error for 65536-byte String16, got nil")
	}
	if buf.Len() != 0 {
		t.Fatalf("expected no bytes written on error, got %d", buf.Len())
	}
}

func TestString16_WriteTo_AcceptsExactlyMax(t *testing.T) {
	s := String16(strings.Repeat("a", 65535))
	var buf bytes.Buffer
	_, err := s.WriteTo(&buf)
	if err != nil {
		t.Fatal(err)
	}
	if buf.Len() != 2+65535 {
		t.Fatalf("want 2+65535 bytes written, got %d", buf.Len())
	}
}
