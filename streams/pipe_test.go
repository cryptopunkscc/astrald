package streams

import (
	"errors"
	"io"
	"strings"
	"testing"
)

func TestWriteAndClose(t *testing.T) {
	left, right := Pipe()

	go func() {
		left.Write([]byte("test"))
		left.Close()
	}()

	var buf = make([]byte, 16)

	n, err := right.Read(buf)
	if n != 4 {
		t.Error("expected 4 bytes, got", n)
	}
	if err != nil {
		t.Error("unexpected error:", err)
	}

	n, err = right.Read(buf)
	if n != 0 {
		t.Error("unexpected data")
	}
	if err != io.EOF {
		t.Error("unexpected error:", err)
	}
}

func TestClose(t *testing.T) {
	var buf = make([]byte, 16)
	left, right := Pipe()

	right.Close()

	_, err := right.Read(buf)
	if err == nil {
		t.Fatal("Read() on a closed pipe should fail")
	}
	if !strings.Contains(err.Error(), "closed pipe") {
		t.Error("Read() on a closed pipe returned an unexpected error:", err)
	}

	_, err = left.Read(buf)
	if err == nil {
		t.Fatal("Read() on a closed pipe should fail")
	}
	if !errors.Is(err, io.EOF) {
		t.Fatal("Read() expected io.EOF error, instead got:", err)
	}
}
