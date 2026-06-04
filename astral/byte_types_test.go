package astral

import (
	"bytes"
	"testing"
)

// Regression tests: Bytes8/16 must reject oversized payloads at WriteTo, not silently
// truncate the length prefix and desync the stream (same lossy-cast shape as String8/16).

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

// TestBytes_MarshalText_PaddingSweep — §8.3. Payload lengths 1..5 cover every base64
// padding case ("==", "=", and no padding) for each Bytes* width.
func TestBytes_MarshalText_PaddingSweep(t *testing.T) {
	for _, payloadLen := range []int{1, 2, 3, 4, 5} {
		payload := make([]byte, payloadLen)
		for i := range payload {
			payload[i] = byte(0xA0 | i)
		}

		t.Run("Bytes8", func(t *testing.T) {
			src := Bytes8(payload)
			text, err := src.MarshalText()
			if err != nil {
				t.Fatal(err)
			}
			var dst Bytes8
			if err := dst.UnmarshalText(text); err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(dst, src) {
				t.Fatalf("Bytes8 padding round-trip mismatch at len=%d: got %x, want %x", payloadLen, dst, src)
			}
		})
		t.Run("Bytes16", func(t *testing.T) {
			src := Bytes16(payload)
			text, err := src.MarshalText()
			if err != nil {
				t.Fatal(err)
			}
			var dst Bytes16
			if err := dst.UnmarshalText(text); err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(dst, src) {
				t.Fatalf("Bytes16 padding round-trip mismatch at len=%d: got %x, want %x", payloadLen, dst, src)
			}
		})
		t.Run("Bytes32", func(t *testing.T) {
			src := Bytes32(payload)
			text, err := src.MarshalText()
			if err != nil {
				t.Fatal(err)
			}
			var dst Bytes32
			if err := dst.UnmarshalText(text); err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(dst, src) {
				t.Fatalf("Bytes32 padding round-trip mismatch at len=%d: got %x, want %x", payloadLen, dst, src)
			}
		})
		t.Run("Bytes64", func(t *testing.T) {
			src := Bytes64(payload)
			text, err := src.MarshalText()
			if err != nil {
				t.Fatal(err)
			}
			var dst Bytes64
			if err := dst.UnmarshalText(text); err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(dst, src) {
				t.Fatalf("Bytes64 padding round-trip mismatch at len=%d: got %x, want %x", payloadLen, dst, src)
			}
		})
	}
}

func TestBytes8_RoundTrip_AtMax(t *testing.T) {
	payload := make([]byte, 255)
	for i := range payload {
		payload[i] = byte(i)
	}
	src := Bytes8(payload)

	var buf bytes.Buffer
	if _, err := src.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}

	var dst Bytes8
	if _, err := dst.ReadFrom(&buf); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(dst, src) {
		t.Fatalf("round-trip mismatch: got %d bytes, want %d", len(dst), len(src))
	}
}
